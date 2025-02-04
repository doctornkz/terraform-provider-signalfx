package signalfx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	detector "github.com/signalfx/signalfx-go/detector"
)

const (
	DetectorAppPath = "/detector/"
)

func detectorResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the detector",
			},
			"program_text": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Signalflow program text for the detector. More info at \"https://developers.signalfx.com/docs/signalflow-overview\"",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the detector",
			},
			"max_delay": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "How long (in seconds) to wait for late datapoints. Max value 900 (15m)",
				ValidateFunc: validation.IntBetween(0, 900),
			},
			"show_data_markers": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "(false by default) When true, markers will be drawn for each datapoint within the visualization.",
			},
			"show_event_lines": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "(false by default) When true, vertical lines will be drawn for each triggered event within the visualization.",
			},
			"disable_sampling": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "(false by default) When false, samples a subset of the output MTS in the visualization.",
			},
			"time_range": &schema.Schema{
				Type:          schema.TypeInt,
				Optional:      true,
				Default:       3600,
				Description:   "Seconds to display in the visualization. This is a rolling range from the current time. Example: 3600 = `-1h`. Defaults to 3600",
				ConflictsWith: []string{"start_time", "end_time"},
			},
			"start_time": &schema.Schema{
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"time_range"},
				Description:   "Seconds since epoch. Used for visualization",
				ValidateFunc:  validation.IntAtLeast(0),
			},
			"end_time": &schema.Schema{
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"time_range"},
				Description:   "Seconds since epoch. Used for visualization",
				ValidateFunc:  validation.IntAtLeast(0),
			},
			"teams": &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Team IDs to associate the detector to",
			},
			"rule": &schema.Schema{
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Set of rules used for alerting",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"severity": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateSeverity,
							Description:  "The severity of the rule, must be one of: Critical, Warning, Major, Minor, Info",
						},
						"detect_label": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "A detect label which matches a detect label within the program text",
						},
						"description": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description of the rule",
						},
						"notifications": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateNotification,
							},
							Description: "List of strings specifying where notifications will be sent when an incident occurs. See https://developers.signalfx.com/v2/docs/detector-model#notifications-models for more info",
						},
						"disabled": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "(default: false) When true, notifications and events will not be generated for the detect label",
						},
						"parameterized_body": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Custom notification message body when an alert is triggered. See https://developers.signalfx.com/v2/reference#detector-model for more info",
						},
						"parameterized_subject": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Custom notification message subject when an alert is triggered. See https://d    evelopers.signalfx.com/v2/reference#detector-model for more info",
						},
						"runbook_url": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "URL of page to consult when an alert is triggered",
						},
						"tip": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Plain text suggested first course of action, such as a command to execute.",
						},
					},
				},
				Set: resourceRuleHash,
			},
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of the detector",
			},
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    timeRangeV0().CoreConfigSchema().ImpliedType(),
				Upgrade: timeRangeStateUpgradeV0,
				Version: 0,
			},
		},

		Create: detectorCreate,
		Read:   detectorRead,
		Update: detectorUpdate,
		Delete: detectorDelete,
		Exists: detectorExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func timeRangeV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"time_range": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func timeRangeStateUpgradeV0(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {

	log.Printf("[DEBUG] SignalFx: Upgrading Detector State %v", rawState["time_range"])
	if tr, ok := rawState["time_range"].(string); ok {
		millis, err := fromRangeToMilliSeconds(tr)
		if err != nil {
			return rawState, err
		}
		rawState["time_range"] = millis / 1000
	}

	return rawState, nil
}

/*
  Use Resource object to construct json payload in order to create a detector
*/
func getPayloadDetector(d *schema.ResourceData) (*detector.CreateUpdateDetectorRequest, error) {

	tfRules := d.Get("rule").(*schema.Set).List()
	rulesList := make([]*detector.Rule, len(tfRules))
	for i, tfRule := range tfRules {
		tfRule := tfRule.(map[string]interface{})
		rule := &detector.Rule{
			Description: tfRule["description"].(string),
			DetectLabel: tfRule["detect_label"].(string),
			Disabled:    tfRule["disabled"].(bool),
		}

		tfSev := tfRule["severity"].(string)
		sev := detector.INFO
		switch tfSev {
		case "Critical":
			sev = detector.CRITICAL
		case "Warning":
			sev = detector.WARNING
		case "Major":
			sev = detector.MAJOR
		case "Minor":
			sev = detector.MINOR
		case "Info":
			sev = detector.INFO
		}
		rule.Severity = sev

		if val, ok := tfRule["parameterized_body"]; ok {
			rule.ParameterizedBody = val.(string)
		}

		if val, ok := tfRule["parameterized_subject"]; ok {
			rule.ParameterizedSubject = val.(string)
		}

		if val, ok := tfRule["runbook_url"]; ok {
			rule.RunbookUrl = val.(string)
		}

		if val, ok := tfRule["tip"]; ok {
			rule.Tip = val.(string)
		}

		if notifications, ok := tfRule["notifications"]; ok {
			notify, err := getNotifications(notifications.([]interface{}))
			if err != nil {
				return nil, err
			}
			rule.Notifications = notify
		}
		rulesList[i] = rule
	}

	cudr := &detector.CreateUpdateDetectorRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		ProgramText: d.Get("program_text").(string),
		Rules:       rulesList,
	}

	if val, ok := d.GetOk("max_delay"); ok {
		maxDelay := int32(val.(int) * 1000)
		cudr.MaxDelay = &maxDelay
	}

	cudr.VisualizationOptions = getVisualizationOptionsDetector(d)

	if val, ok := d.GetOk("teams"); ok {
		teams := []string{}
		for _, t := range val.([]interface{}) {
			teams = append(teams, t.(string))
		}
		cudr.Teams = teams
	}

	return cudr, nil
}

func getVisualizationOptionsDetector(d *schema.ResourceData) *detector.Visualization {
	viz := detector.Visualization{}

	if val, ok := d.GetOk("show_data_markers"); ok {
		viz.ShowDataMarkers = val.(bool)
	}
	if val, ok := d.GetOk("show_event_lines"); ok {
		viz.ShowEventLines = val.(bool)
	}
	if val, ok := d.GetOk("disable_sampling"); ok {
		viz.DisableSampling = val.(bool)
	}

	if val, ok := d.GetOk("time_range"); ok {
		tr := &detector.Time{}
		r := int64(val.(int)) * 1000
		tr.Range = &r
		tr.Type = "relative"
		viz.Time = tr
	}
	if val, ok := d.GetOk("start_time"); ok {
		tr := &detector.Time{}
		tr.Type = "absolute"
		start := val.(int64) * 1000
		tr.Start = &start
		if val, ok := d.GetOk("end_time"); ok {
			end := val.(int64) * 1000
			tr.End = &end
		}
		viz.Time = tr
	}

	if (detector.Visualization{}) == viz {
		// Return a nil ptr so we don't serialize nothing
		return nil
	}

	return &viz
}

func detectorCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalfxConfig)
	payload, err := getPayloadDetector(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}

	debugOutput, _ := json.Marshal(payload)
	log.Printf("[DEBUG] SignalFx: Create Detector Payload: %s", string(debugOutput))

	det, err := config.Client.CreateDetector(payload)
	if err != nil {
		return err
	}
	// Since things worked, set the URL and move on
	appURL, err := buildAppURL(config.CustomAppURL, DetectorAppPath+det.Id)
	if err != nil {
		return err
	}
	if err := d.Set("url", appURL); err != nil {
		return err
	}
	d.SetId(det.Id)

	return detectorAPIToTF(d, det)
}

func detectorExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	config := meta.(*signalfxConfig)
	_, err := config.Client.GetDetector(d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func detectorRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalfxConfig)
	det, err := config.Client.GetDetector(d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			d.SetId("")
		}
		return err
	}

	appURL, err := buildAppURL(config.CustomAppURL, DetectorAppPath+det.Id)
	if err != nil {
		return err
	}
	if err := d.Set("url", appURL); err != nil {
		return err
	}

	return detectorAPIToTF(d, det)
}

func detectorAPIToTF(d *schema.ResourceData, det *detector.Detector) error {
	debugOutput, _ := json.Marshal(det)
	log.Printf("[DEBUG] SignalFx: Got Detector to enState: %s", string(debugOutput))

	if err := d.Set("name", det.Name); err != nil {
		return err
	}
	if err := d.Set("description", det.Description); err != nil {
		return err
	}
	if err := d.Set("program_text", det.ProgramText); err != nil {
		return err
	}
	// We divide by 1000 because the API uses millis, but this provider uses
	// seconds
	if det.MaxDelay != nil {
		if err := d.Set("max_delay", *det.MaxDelay/1000); err != nil {
			return err
		}
	}
	if err := d.Set("teams", det.Teams); err != nil {
		return err
	}
	viz := det.VisualizationOptions
	if viz != nil {
		if err := d.Set("show_data_markers", viz.ShowDataMarkers); err != nil {
			return err
		}
		if err := d.Set("show_event_lines", viz.ShowEventLines); err != nil {
			return err
		}
		if err := d.Set("disable_sampling", viz.DisableSampling); err != nil {
			return err
		}

		tr := viz.Time
		if tr != nil {
			// We divide by 1000 because the API uses millis, but this provider uses
			// seconds
			if tr.Range != nil {
				if err := d.Set("time_range", *tr.Range/1000); err != nil {
					return err
				}
			}
			if err := d.Set("start_time", tr.Start); err != nil {
				return err
			}
			if err := d.Set("end_time", tr.End); err != nil {
				return err
			}
		}
	}

	rules := make([]map[string]interface{}, len(det.Rules))
	for i, r := range det.Rules {
		rule := make(map[string]interface{})
		rule["severity"] = r.Severity
		rule["detect_label"] = r.DetectLabel
		rule["description"] = r.Description

		notifications := make([]string, len(r.Notifications))
		for i, not := range r.Notifications {
			tfNot, err := getNotifyStringFromAPI(not)
			if err != nil {
				return err
			}
			notifications[i] = tfNot
		}
		rule["notifications"] = notifications
		rule["disabled"] = r.Disabled
		rule["parameterized_body"] = r.ParameterizedBody
		rule["parameterized_subject"] = r.ParameterizedSubject
		rule["runbook_url"] = r.RunbookUrl
		rule["tip"] = r.Tip
		rules[i] = rule
	}
	if err := d.Set("rule", rules); err != nil {
		return err
	}

	return nil
}

func detectorUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalfxConfig)
	payload, err := getPayloadDetector(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}

	debugOutput, _ := json.Marshal(payload)
	log.Printf("[DEBUG] SignalFx: Update Detector Payload: %s", string(debugOutput))

	det, err := config.Client.UpdateDetector(d.Id(), payload)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] SignalFx: Update Detector Response: %v", det)
	// Since things worked, set the URL and move on
	appURL, err := buildAppURL(config.CustomAppURL, DetectorAppPath+det.Id)
	if err != nil {
		return err
	}
	if err := d.Set("url", appURL); err != nil {
		return err
	}
	d.SetId(det.Id)
	return detectorAPIToTF(d, det)
}

func detectorDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalfxConfig)

	return config.Client.DeleteDetector(d.Id())
}

/*
   Hashing function for rule substructure of the detector resource, used in determining state changes.
*/
func resourceRuleHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["description"]))
	buf.WriteString(fmt.Sprintf("%s-", m["severity"]))
	buf.WriteString(fmt.Sprintf("%s-", m["detect_label"]))
	buf.WriteString(fmt.Sprintf("%s-", m["disabled"]))

	// loop through optional rule attributes
	var optional_rule_keys = []string{"parameterized_body", "parameterized_subject", "runbook_url", "tip"}

	for _, key := range optional_rule_keys {
		if val, ok := m[key]; ok {
			buf.WriteString(fmt.Sprintf("%s-", val))
		}
	}

	// Sort the notifications so that we generate a consistent hash
	if v, ok := m["notifications"]; ok {
		notifications := v.([]interface{})
		s_notifications := make([]string, len(notifications))
		for i, raw := range notifications {
			s_notifications[i] = raw.(string)
		}
		sort.Strings(s_notifications)

		for _, notification := range s_notifications {
			buf.WriteString(fmt.Sprintf("%s-", notification))
		}
	}

	return hashcode.String(buf.String())
}

/*
  Validates the severity field against a list of allowed words.
*/
func validateSeverity(v interface{}, k string) (we []string, errors []error) {
	value := v.(string)
	allowedWords := []string{"Critical", "Major", "Minor", "Warning", "Info"}
	for _, word := range allowedWords {
		if value == word {
			return
		}
	}
	errors = append(errors, fmt.Errorf("%s not allowed; must be one of: %s", value, strings.Join(allowedWords, ", ")))
	return
}
