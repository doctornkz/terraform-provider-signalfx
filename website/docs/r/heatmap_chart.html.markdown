---
layout: "signalfx"
page_title: "SignalFx: signalfx_heatmap_chart"
sidebar_current: "docs-signalfx-resource-heatmap-chart"
description: |-
  Allows Terraform to create and manage SignalFx heat map chart
---

# Resource: signalfx_heatmap_chart

This chart type displays the specified plot in a heatmap fashion. This format is similar to the [Infrastructure Navigator](https://signalfx-product-docs.readthedocs-hosted.com/en/latest/built-in-content/infra-nav.html#infra), with squares representing each source for the selected metric, and the color of each square representing the value range of the metric.

## Example Usage

```terraform
resource "signalfx_heatmap_chart" "myheatmapchart0" {
    name = "CPU Total Idle - Heatmap"

    program_text = <<-EOF
        myfilters = filter("cluster_name", "prod") and filter("role", "search")
        data("cpu.total.idle", filter=myfilters).publish()
        EOF

    description = "Very cool Heatmap"

    disable_sampling = true
    sort_by = "+host"
    group_by = ["hostname", "host"]
    hide_timestamp = true
}
```


## Argument Reference

The following arguments are supported in the resource block:

* `name` - (Required) Name of the chart.
* `program_text` - (Required) Signalflow program text for the chart. More info at <https://developers.signalfx.com/docs/signalflow-overview>.
* `description` - (Optional) Description of the chart.
* `unit_prefix` - (Optional) Must be `"Metric"` or `"Binary`". `"Metric"` by default.
* `minimum_resolution` - (Optional) The minimum resolution (in seconds) to use for computing the underlying program.
* `max_delay` - (Optional) How long (in seconds) to wait for late datapoints.
* `refresh_interval` - (Optional) How often (in seconds) to refresh the values of the heatmap.
* `disable_sampling` - (Optional) If `false`, samples a subset of the output MTS, which improves UI performance. `false` by default.
* `group_by` - (Optional) Properties to group by in the heatmap (in nesting order).
* `sort_by` - (Optional) The property to use when sorting the elements. Must be prepended with `+` for ascending or `-` for descending (e.g. `-foo`).
* `hide_timestamp` - (Optional) Whether to show the timestamp in the chart. `false` by default.
* `color_range` - (Required, if `color_scale` is not used) Values and color for the color range. Example: `color_range : { min : 0, max : 100, color : "#0000ff" }`. Look at this [link](https://docs.signalfx.com/en/latest/charts/chart-options-tab.html).
    * `min_value` - (Optional) The minimum value within the coloring range.
    * `max_value` - (Optional) The maximum value within the coloring range.
    * `color` - (Required) The color range to use. The starting hex color value for data values in a heatmap chart. Specify the value as a 6-character hexadecimal value preceded by the '#' character, for example "#ea1849" (grass green).
* `color_scale` - (Required, if `color_range` is not used) Single color range including both the color to display for that range and the borders of the range. Example: `[{ gt = 60, color = "blue" }, { lte = 60, color = "yellow" }]`. Look at this [link](https://docs.signalfx.com/en/latest/charts/chart-options-tab.html).
    * `gt` - (Optional) Indicates the lower threshold non-inclusive value for this range.
    * `gte` - (Optional) Indicates the lower threshold inclusive value for this range.
    * `lt` - (Optional) Indicates the upper threshold non-inculsive value for this range.
    * `lte` - (Optional) Indicates the upper threshold inclusive value for this range.
    * `color` - (Required) The color range to use. Must be either gray, blue, navy, orange, yellow, magenta, purple, violet, lilac, green, aquamarine.
