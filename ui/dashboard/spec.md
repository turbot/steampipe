# Mods & Reports

## Resource Types

You can compose reports using a number of resource types.

| Name        | Purpose                                                                                                                                                           | Valid children types                                                                                               | Allowed at top-level |
| ----------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ | -------------------- |
| `report`    | A way to compose resources to meet a reporting requirement. Within the Steampipe reporting UIs, each `report` will be presented as an available item to run.      | `chart`<br/>`container`<br/>`control`<br/>`counter`<br/>`hierarchy`<br/>`image`<br/>`input`<br/>`table`<br/>`text` | Yes                  |
| `container` | Conceptually identical to a report, except it will not be presented as an available item to run within the Steampipe reporting UIs.                               | `chart`<br/>`container`<br/>`control`<br/>`counter`<br/>`hierarchy`<br/>`image`<br/>`input`<br/>`table`<br/>`text` | Yes                  |
| `chart`     | Chart SQL data in a number of ways e.g. `bar`, `column`, `line`, `pie` etc.                                                                                       | None                                                                                                               | Yes                  |
| `control`   | Allow mod controls to be defined or re-used in reports.                                                                                                           | None                                                                                                               | Yes                  |
| `counter`   | Allow anything from a simple value to change in value to be presented in different styles e.g. `plain`, `alert` etc. Supports static values, or derived from SQL. | None                                                                                                               | Yes                  |
| `hierarchy` | Visualise hierarchical data using things such as `sankey` or `graph`.                                                                                             | None                                                                                                               | Yes                  |
| `image`     | Allow embedding of images in reports. Supports static URLs, or can be derived from SQL.                                                                           | None                                                                                                               | Yes                  |
| `input`     | Allow dynamic reporting based on user-provided `input`.                                                                                                           | None                                                                                                               | Yes                  |
| `table`     | Show tabular data in a report.                                                                                                                                    | None                                                                                                               | Yes                  |
| `text`      | Allow GitHub-flavoured markdown to be added to a report.                                                                                                          | None                                                                                                               | Yes                  |
| more TBD... |                                                                                                                                                                   |                                                                                                                    |                      |

## Spec

### `report`

A report is designed to be used by consumers of your mod to answer specific reporting questions, such as "How many public AWS buckets do I have?" or "Show me the number of aging Zendesk tickets by owner".

Reports can only be declared at the top-level of a mod and can be re-used by using a `container`.

For layout, a report consists of 12 grid units, where items inside it will consume the full 12 grid units, unless they specific an explicit [width](#width).

#### Properties

- `title`: Plain text title used to display in lists, page title etc. When viewing the report in a browser, will be rendered as a h1.

#### Examples

```hcl
report hello_world {
  text {
    value = "# Hello world!"
  }
}
```

```hcl
# Existing report re-used with `base`
report compose_other {
  container {
    base = report.hello_world
  }
}
```

### `container`

A container is intended to be used to group/lay-out related items within a report. For example, you may want to group together a number of different charts for a specific AWS service.

Containers can be declared as named resources at the top level of a mod, or be declared as anonymous blocks inside a `report` or `container`, or be re-used inside a `report` or `container` by using a `container` with `base = <mod>.container.<container_resource_name>`.

#### Properties

- `base`: A reference to a named `report` or `container` that this `container` should source its definition from. `title` and `width` can be overridden after sourcing via `base`.
- `title`: A container title is a shorthand convenience for adding a `text` block with a h2 Markdown heading and wrapping both items in a container. E.g. a title of `AWS S3 Metrics` would end up as:

```hcl
container {
  text {
    value = "## AWS S3 Metrics"
  }

  container {
    # Original container containing the title here...
  }
}
```

- `width`: See [width](#width).

#### Examples

```hcl
# Simple container with a text markdown panel

container {
  text {
    value = "## AWS S3 Metrics"
  }
}
```

```hcl
# Full-width container with 2 side-by-side containers on larger screens.

container {
  container {
    width = 6
    text {
      value = "## First text"
    }
  }

  container {
    width = 6
    text {
      value = "## Second text"
    }
  }
}
```

```hcl
# Re-use an existing named top-level container and override its width.

container {
  base  = container.some_other_container
  width = 6
}
```

### `chart`

A chart will allow visualisation of queries in a variety of charting types such as `bar`, `column`, `donut`, `line` or `pie`.

The chart types share key things such as shape of the date and common configuration, so by having a single `chart` type, this allows us to be flexible. For example, you could simply change the type of chart from `bar` to `line` and it will just work, or we could allow in the reporting UI for a user to switch the chart type - all made possible by having a consistent data structure and commonality across the charts.

Chart blocks can be declared as named resources at the top level of a mod, or be declared as anonymous blocks inside a `report` or `container`, or be re-used inside a `report` or `container` by using a `chart` with `base = <mod>.chart.<chart_resource_name>`.

#### Properties

- `base`: A reference to a named `chart` resource that this `chart` should source its definition from. `title` and `width` can be overridden after sourcing via `base`.
- `title`: As per the title for a `container` or `chart`, except as this is a leaf is rendered as h3.
- `type`: The type of the chart. Can be `bar`, `column`, `donut`, `line` or `pie`.
- `sql`: Inline SQL or named query.
- `width`: See [width](#width).
- `legend`: See [legend](#legend).
- `series`: A named block matching the name of the series you wish to configure. See `Series`.
- `axes`: See [axes](#axes).

#### Data Format

| X-Axis  | Y-Axis Series 1  | Y-Axis Series 2  | ... | Y-Axis Series N  |
| ------- | ---------------- | ---------------- | --- | ---------------- |
| Label 1 | Value 1 Series 1 | Value 1 Series 2 | ... | Value 1 Series N |
| Label 2 | Value 2 Series 1 | Value 2 Series 2 | ... | Value 2 Series N |
| ...     | ...              | ...              | ... | ...              |
| Label N | Value N Series 1 | Value 1 Series 2 | ... | Value N Series N |

#### Examples

For the chart examples, we will be referring to the following single and multi-series queries:

```hcl
query aws_a3_buckets_by_region {
  sql = <<EOF
select
    region as Region,
    count(*) as Total
from
    aws_s3_bucket
group by
    region
order by
    Total desc
EOF
}
```

```hcl
query aws_a3_unencrypted_and_nonversioned_buckets_by_region {
  sql = <<EOF
with unencrypted_buckets_by_region as (
  select
    region,
    count(*) as unencrypted
  from
    aws_morales_aaa.aws_s3_bucket
  where
    server_side_encryption_configuration is null
  group by
    region
),
nonversioned_buckets_by_region as (
  select
    region,
    count(*) as nonversioned
  from
    aws_morales_aaa.aws_s3_bucket
  where
    not versioning_enabled
  group by
    region
),
other_buckets_by_region as (
  select
    region,
    count(*) as "other"
  from
    aws_morales_aaa.aws_s3_bucket
  where
    server_side_encryption_configuration is not null
    and versioning_enabled
  group by
    region
)
select
  o.region,
  o.other,
  u.unencrypted,
  v.nonversioned
from
  other_buckets_by_region o
  full join unencrypted_buckets_by_region u on o.region = u.region
  full join nonversioned_buckets_by_region v on o.region = v.region;
EOF
}
```

##### Bar

```hcl
# Single-series bar chart accepting all of the defaults

chart {
  type = "bar"
  sql = query.aws_a3_buckets_by_region.sql
  title = "AWS S3 Buckets by Region"
}
```

```hcl
# Multi-series bar chart providing custom properties for a variety of chart features

chart {
  type = "bar"
  sql = query.aws_a3_unencrypted_and_nonversioned_buckets_by_region.sql
  title = "Unencrypted and Non-Versioned Buckets by Region"
  legend {
    display  = "auto"
    position = "top"
  }
  series other {
    title = "Configured Buckets"
    color = "green"
  }
  series unencrypted {
    title = "Unencrypted Buckets"
    color = "red"
  }
  series nonversioned {
    title = "Non-Versioned Buckets"
    color = "orange"
  }
  axes {
    x {
      title = "Regions"
      labels {
        display = "auto"
      }
    }
    y {
      title  = "Totals"
      labels {
        display = "show"
      }
      min    = 0
      max    = 100
      steps  = 10
    }
  }
}
```

##### Column

```hcl
# Single-series column chart accepting all of the defaults

chart {
  type = "column"
  sql = query.aws_a3_buckets_by_region.sql
  title = "AWS S3 Buckets by Region"
}
```

##### Donut

```hcl
# Single-series donut chart accepting all of the defaults

chart {
  type = "donut"
  sql = query.aws_a3_buckets_by_region.sql
  title = "AWS S3 Buckets by Region"
}
```

##### Line

```hcl
# Single-series line chart accepting all of the defaults

chart {
  type = "line"
  sql = query.aws_a3_buckets_by_region.sql
  title = "AWS S3 Buckets by Region"
}
```

##### Pie

```hcl
# Single-series pie chart accepting all of the defaults

chart {
  type = "pie"
  sql = query.aws_a3_buckets_by_region.sql
  title = "AWS S3 Buckets by Region"
}
```

### `counter`

A counter is used to show a value to the user, or some change in value. A counter can also present itself in different styles e.g. show me a count of public S3 buckets and if the value is greater than 0 show as an `alert`, else as `ok`.

Counters can be declared as named resources at the top level of a mod, or be declared as anonymous blocks inside a `report` or `container`, or be re-used inside a `report` or `container` by using a `counter` with `base = <mod>.counter.<counter_resource_name>`.

#### Properties

- `base`: A reference to a named `counter` resource that this `counter` should source its definition from. `label`, `title`, `value`, `style` and `width` can be overridden after sourcing via `base`.
- `title`: As per the title for a `container` or `chart`, except as this is a leaf is rendered as h3.
- `label`: Inferred from the first column name in simple data format. Else can be set explicitly in HCL, or returned by the query in the `label` column in the formal data format.
- `sql`: Inline SQL or named query.
- `value`: Inferred from the first column's value in simple data format.
- `style`: `plain` (default), `alert`, `info` or `ok`.
- `width`: See [width](#width).

#### Data Structure

A counter supports 2 data structures.

1. A simple structure where column 1's name is the counter `label` and column 1's value is the counter `value`.
2. A formal data structure where the column names map to properties of the `counter`.

Simple data structure:

| \<label> |
| -------- |
| \<value> |

Formal data structure:

| label               | value | style |
| ------------------- | ----- | ----- |
| Unencrypted Buckets | 10    | alert |

#### Examples

```hcl
# Hard-code value and style declared in HCL. Useful for mocking report layouts.

counter {
  title = "Mock Value"
  value = 10
  style = "alert"
}
```

```hcl
# Using simple data structure as above. Style is static.

counter {
  sql = <<EOF
select
  count(*) as "Unencrypted Buckets"
from
  aws_s3_bucket
where
  server_side_encryption_configuration is null;
EOF
  style = "alert"
}
```

```hcl
# Using formal data structure as above, where the query determines properties of the counter. E.g. style is derived from the `style` column of the query.

counter {
  sql = <<EOF
select
  'Unencrypted Buckets' as title,
  count(*) as value,
  case
    when count(*) > 0 then 'alert'
    else 'ok'
  end style
from
  aws_s3_bucket
where
  server_side_encryption_configuration is null;
EOF
}
```

### `image`

Display images in reports., either from a known publicly-accessible `src` URL, or from a `src` return in a SQL query.

Image blocks can be declared as named resources at the top level of a mod, or be declared as anonymous blocks inside a `report` or `container`, or be re-used inside a `report` or `container` by using an `image` with `base = <mod>.image.<image_resource_name>`.

#### Properties

- `base`: A reference to a named `text` resource that this `text` should source its definition from. `title` and `width` can be overridden after sourcing via `base`.
- `title`: As per the title for a `container` or `chart`, except as this is a leaf is rendered as h3.
- `src`: Publicly-accessible URL for the image.
- `alt`: Alternative text for the image.
- `width`: See [width](#width).

#### Data Structure

If an image provides a `sql` query, it supports 2 data structures.

1. A simple structure where column 1's name is the `alt` and column 1's value is the image `src`.
2. A formal data structure where the column names map to properties of the `image`.

Simple data structure:

| \<alt> |
| ------ |
| \<src> |

Formal data structure:

| src                                  | alt            |
| ------------------------------------ | -------------- |
| https://steampipe.io/images/logo.png | Steampipe Logo |

#### Examples

```hcl
# Hard-code src and alt declared in HCL. Useful for known image URLs.

image {
  src = "https://steampipe.io/images/logo.png"
  alt = "Steampipe Logo"
}
```

```hcl
# Using simple data structure as above.

image {
  sql = <<EOF
select
  favicon as "Steampipe Logo"
from
  web
where
  domain = "https://steampipe.io";
EOF
}
```

```hcl
# Using formal data structure as above, where the query determines properties of the image. E.g. src is derived from the `src` column of the query.

image {
  sql = <<EOF
select
  favicon as src,
  "Steampipe Logo" as alt
from
  web
where
  domain = "https://steampipe.io";
EOF
}
```

### `input`

Allow user input in reports using common form components such as inputs, selects etc. Data can either be static or derived from a SQL query.

The value of an input is synced with the declared `variable`, which allows the default/initial value to be set by the `variable` and then allows other mod resources to consume this `variable` to enable dynamic behaviour.

Input blocks can be declared as named resources at the top level of a mod, or be declared as anonymous blocks inside a `report` or `container`, or be re-used inside a `report` or `container` by using an `input` with `base = <mod>.input.<input_resource_name>`.

#### Properties

- `base`: A reference to a named `input` resource that this `input` should source its definition from. `title`, `sql`, `type`, `options` and `width` can be overridden after sourcing via `base`.
- `title`: As per the title for a `container` or `chart`, except as this is a leaf is rendered as h3.
- `sql`: Inline SQL or named query.
- `type`: The type of the input. Can be `check`, `input`, `radio` or `select`.
- `variable`: A reference to a mod `variable` that backs the input.
- `options`: Only applicable for input types that display options e.g. `check`, `radio` or `select` that wish to display a static list of options. Inferred from the data first 2 columns value in simple data format.
- `placeholder`: Only applicable for `input` type.
- `width`: See [width](#width).

#### Data Structure

The data structure for an `input` will depend on the `type`.

##### Check / Radio / Select

| label               | value                 |
| ------------------- | --------------------- |
| default             | vpc-05657e5bef9676266 |
| acme @ 10.84.0.0/16 | vpc-03656e5eef967f366 |

#### Examples

```hcl
# Select with static options

input {
  type = "select"
  variable = var.selected_vpc
  options = [
    {
      label = "default"
      value = "vpc-05657e5bef9676266"
    }
    {
      label = "acme @ 10.84.0.0/16"
      value = "vpc-03656e5eef967f366"
    }
  ]
}
```

```hcl
# Select with dynamic options

input {
  type = "select"
  variable = var.selected_vpc
  sql = <<EOF
select
  title as label
  vpc_id as value,
from
  aws_vpc;
EOF
}
```

```hcl
# Radio with static options

input {
  type = "radio"
  variable = var.selected_vpc
  options = [
    {
      label = "default"
      value = "vpc-05657e5bef9676266"
    }
    {
      label = "acme @ 10.84.0.0/16"
      value = "vpc-03656e5eef967f366"
    }
  ]
}
```

```hcl
# Radio with dynamic options

input {
  type = "radio"
  variable = var.selected_vpc
  sql = <<EOF
select
  title as label
  vpc_id as value,
from
  aws_vpc;
EOF
}
```

```hcl
# Check with static options

input {
  type = "check"
  variable = var.selected_vpc
  options = [
    {
      label = "default"
      value = "vpc-05657e5bef9676266"
    }
    {
      label = "acme @ 10.84.0.0/16"
      value = "vpc-03656e5eef967f366"
    }
  ]
}
```

```hcl
# Check with dynamic options

input {
  type = "check"
  variable = var.selected_vpc
  sql = <<EOF
select
  title as label
  vpc_id as value,
from
  aws_vpc;
EOF
}
```

### `text`

Display either rendered `markdown` ([GitHub Flavored Markdown](https://github.github.com/gfm/)) or `raw` text with no interpretation of markup. # TODO do we want HTML? How can we be sure it's safe before rendering in the browser?

Text blocks can be declared as named resources at the top level of a mod, or be declared as anonymous blocks inside a `report` or `container`, or be re-used inside a `report` or `container` by using a `text` with `base = <mod>.text.<text_resource_name>`.

#### Properties

- `base`: A reference to a named `text` resource that this `text` should source its definition from. `title` and `width` can be overridden after sourcing via `base`.
- `title`: As per the title for a `container` or `chart`, except as this is a leaf is rendered as h3.
- `type`: `markdown` (default) or `raw`.
- `value`: The `markdown` or `html` string to use. Can also be sourced using the HCL `file` function.
- `width`: See [width](#width).

#### Examples

```hcl
# Render as markdown (default).

text {
  value = <<EOF
# I am some markdown text.

Please respect my markdown.
EOF
}
```

```hcl
# Explicitly specify markdown (but is optional as markdown is the default).

text {
  type  = "markdown"
  value = <<EOF
# I am some markdown text.

Please respect my markdown.
EOF
}
```

```hcl
# Render raw

text {
  type  = "raw"
  value = "<h2>I am a HTML title, but I'll be displayed as-is</h2>"
}
```

### Common Properties:

#### `title`

Optional `title` for an item. Provided as plain text, but will be rendered as a `text` panel with a `style` of `markdown` using h1 (for `report`), h2 (for `container`) or h3 (for any leaf nodes e.g. `chart`). This `text` block and the item it is titling will be wrapped by a `container`.

#### `width`

The number of grid units that this item should consume from its parent. In the browser, this is only used when in a large screen size and up (viewport min-width of 1024px), and will always be 12 for smaller viewports. Provided value must be between `0` and `12` (default). `0` would mean the `container` is hidden and will not consume any space on the report.

### Common Chart Properties

#### `axes`

Applicable to `bar`, `column`, `line` and `scatter`.

- `x`:

| Property | Type                         | Default | Values | Description |
| -------- | ---------------------------- | ------- | ------ | ----------- |
| `title`  | See [axistitle](#axis-title) |         |        |             |
| `labels` | See [labels](#labels)        |         |        |             |

- `y`:

| Property | Type                         | Default                                                                                                                                                                   | Values            | Description |
| -------- | ---------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------- | ----------- |
| `title`  | See [axistitle](#axis-title) |                                                                                                                                                                           |                   |             |
| `labels` | See [labels](#labels)        |                                                                                                                                                                           |                   |             |
| `min`    | number                       | Determined by the range of values. For positive ranges, this will be `0`. For negative ranges, this will be scaled to the next appropriate value below the range min.     | Any valid number. |             |
| `max`    | number                       | Determined by the range of values. For positive ranges, this will be scaled to the next appropriate value above the range max `0`. For negative ranges, this will be `0`. | Any valid number. |             |

#### `axis title`

| Property  | Type   | Default  | Values                                | Description                                                                                                                              |
| --------- | ------ | -------- | ------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `display` | string | `none`   | `always` or `none` (default).         | `always` will ensure the axis title is `always` shown, or `none` will never show it.                                                     |
| `align`   | string | `center` | `start`, `center` (default) or `end`. | By default the chart will align the axis title in the `center` of the chart, but this can be overridden to `start` or `end` if required. |
| `value`   | string |          | Max 50 characters.                    |                                                                                                                                          |

#### `labels`

| Property  | Type   | Default | Values                                | Description                                                                                                                                                                                                        |
| --------- | ------ | ------- | ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `display` | string | `auto`  | `auto` (default), `always` or `none`. | `auto` will display as many labels as possible for the size of the chart. `always` will always show all labels, but will truncate them with an ellipsis as necessary. `none` will never show labels for this axis. |
| `format`  | TBD    |         |                                       |                                                                                                                                                                                                                    |

#### `legend`

| Property   | Type   | Default | Values                              | Description                                                                                                                                        |
| ---------- | ------ | ------- | ----------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `display`  | string | `auto`  | `auto`, `always` or `none`.         | `auto` will display a legend if there are multiple data series. `show` will ensure a legend is `always` shown, or `hide` will never show a legend. |
| `position` | string | `top`   | `top`, `right`, `bottom` or `left`. | By default the chart will display a legend at the `top` of the chart, but this can be overridden to `right`, `bottom` or `left` if required.       |

#### `series`:

| Property | Type   | Default                                                              | Values                                                                                                                                  | Description |
| -------- | ------ | -------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- | ----------- |
| `title`  | string | The column name that the series data resides in.                     | Max 50 characters                                                                                                                       |             |
| `color`  | string | The matching color from the default theme for the data series index. | Any valid color from the HTML extended colors (e.g. `gold`), a valid Hex string (e.g. `#0000FF`) or a valid RGB (e.g. `rgb(0, 0, 100)`) |             |

---

Inputs mock:

```hcl
query aws_regions_input {
  sql = <<-EOT
select
  name as value,
  title as label
from
  aws_region;
EOT
}

query aws_vpcs_for_region_input {
  sql = <<-EOT
select
  vpc_id as value,
  title as label
from
  aws_morales_aaa.aws_vpc
where
  region = $1;
EOT
  param region {
    description = "Region to filter VPCs by."
  }
}

execute
{
  region = "us-east-1"
}

report inputs_mock {
  title = "Inputs Mock"

  param "region" {
    default = var.default_region
    sql = query.aws_regions_input.sql
    type = "select"
    description = "Selected region for the report."
  }

  param "region" {
    default = var.default_region
    description = "Selected region for the report."
  }

  param "region" {
    default = var.default_region
    value = input.region_select.value
    description = "Selected region for the report."
  }

  param "vpc_id" {
    default = input.vpc_select.value
    description = "Selected VPC for the report."
  }

  container {
    input "foo" {
      base = param.region
      width = 6
    }

    input {
      base = param.region
      sql = query.aws_regions_input.sql
      type = "select"
      width = 6
    }

    input "region_select" {
      type = "select"
      sql = query.aws_regions_input.sql
      width = 6
    }

    input {
      base = param.region
      type = "select"
      sql = query.aws_regions_input.sql
      width = 6
    }

    input "vpc_select" {
      type = "select"
      query = query.aws_vpcs_for_region_input
      args = {
        region = param.region
      }
      width = 6
    }
  }

  container {
    chart {
      type = "bar"
      # sql = query.foo.sql
      query = query.aws_blah_for_vpc
      args = {
        region = param.vpc_id
      }
    }
  }

  text {
    type = "markdown"
    value = "## Some text"
  }
}
```

## OLD DOCS - ignore below here

### `bar`

Display a `bar` chart for single or grouped multi-series. Maximum series count is `10`.

#### Data Format

| X-Axis  | Y-Axis Series 1  | Y-Axis Series 2  | ... | Y-Axis Series N  |
| ------- | ---------------- | ---------------- | --- | ---------------- |
| Label 1 | Value 1 Series 1 | Value 1 Series 2 | ... | Value 1 Series N |
| Label 2 | Value 2 Series 1 | Value 2 Series 2 | ... | Value 2 Series N |
| ...     | ...              | ...              | ... | ...              |
| Label N | Value N Series 1 | Value 1 Series 2 | ... | Value N Series N |

#### Properties

- `title`: See [title](#title).
- `sql`: Inline SQL or named query.
- `legend`: See [legend](#legend).
- `series`: A named block matching the name of the series you wish to configure. See `Series`.
- `axes`: See [axes](#axes).

##### Basic

Accept all defaults.

```hcl
bar {
  sql = query.aws_a3_buckets_by_region.sql
  title = "AWS S3 Buckets by Region"
}
```

##### Kitchen Sink

Specify all options.

# TODO should blocks assign with =? E.g. `legend { ...` vs `legend = { ...`

```hcl
bar {
  sql = query.aws_a3_unencrypted_and_nonversioned_buckets_by_region.sql
  title = "Unencrypted and Non-Versioned Buckets by Region"
  legend {
    display  = "auto"
    position = "top"
  }
  series other {
    title = "Configured Buckets"
    color = "green"
  }
  series unencrypted {
    title = "Unencrypted Buckets"
    color = "red"
  }
  series nonversioned {
    title = "Non-Versioned Buckets"
    color = "orange"
  }
  axes {
    x {
      title = "Regions"
      labels {
        display = "auto"
      }
    }
    y {
      title  = "Totals"
      labels {
        display = "show"
      }
      min    = 0
      max    = 100
      steps  = 10
    }
  }
}
```

### `column`

Display a `column` chart for single or grouped multi-series. Maximum series count is 10.

#### Data Format

| X-Axis  | Y-Axis Series 1  | Y-Axis Series 2  | ... | Y-Axis Series N  |
| ------- | ---------------- | ---------------- | --- | ---------------- |
| Label 1 | Value 1 Series 1 | Value 1 Series 2 | ... | Value 1 Series N |
| Label 2 | Value 2 Series 1 | Value 2 Series 2 | ... | Value 2 Series N |
| ...     | ...              | ...              | ... | ...              |
| Label N | Value N Series 1 | Value 1 Series 2 | ... | Value N Series N |

#### Properties

- `title`: See [title](#title).
- `sql`: Inline SQL or named query.
- `legend`: See [legend](#legend).
- `series`: A named block matching the name of the series you wish to configure. See `Series`.
- `axes`: See [axes](#axes).

##### Basic

Accept all defaults.

```hcl
column {
  sql = query.aws_a3_buckets_by_region.sql
  title = "AWS S3 Buckets by Region"
}
```

##### Kitchen Sink

Specify all options.

```hcl
column {
  sql = query.aws_a3_unencrypted_and_nonversioned_buckets_by_region.sql
  title = "Unencrypted and Non-Versioned Buckets by Region"
  legend {
    display  = "auto"
    position = "top"
  }
  series other {
    title = "Configured Buckets"
    color = "green"
  }
  series unencrypted {
    title = "Unencrypted Buckets"
    color = "red"
  }
  series nonversioned {
    title = "Non-Versioned Buckets"
    color = "orange"
  }
  axes {
    x {
      title = "Regions"
      labels = "auto"
    }
    y {
      title  = "Totals"
      labels = {
        display = "show"
      }
      min    = 0
      max    = 100
      steps  = 10
    }
  }
}
```

### `line`

Display a `line` chart for single or multi-series. Maximum series count is 10.

#### Data Format

| X-Axis  | Y-Axis Series 1  | Y-Axis Series 2  | ... | Y-Axis Series N  |
| ------- | ---------------- | ---------------- | --- | ---------------- |
| Label 1 | Value 1 Series 1 | Value 1 Series 2 | ... | Value 1 Series N |
| Label 2 | Value 2 Series 1 | Value 2 Series 2 | ... | Value 2 Series N |
| ...     | ...              | ...              | ... | ...              |
| Label N | Value N Series 1 | Value 1 Series 2 | ... | Value N Series N |

#### Properties

- `title`: See [title](#title).
- `sql`: Inline SQL or named query.
- `legend`: See [legend](#legend).
- `series`: A named block matching the name of the series you wish to configure. See `Series`.
- `axes`: See [axes](#axes).

##### Basic

Accept all defaults.

```hcl
column {
  sql = query.aws_a3_buckets_by_region.sql
  title = "AWS S3 Buckets by Region"
}
```

##### Kitchen Sink

Specify all options.

```hcl
line {
  sql = query.aws_a3_unencrypted_and_nonversioned_buckets_by_region.sql
  title = "Unencrypted and Non-Versioned Buckets by Region"
  legend = {
    display  = "auto"
    position = "top"
  }
  series other {
    title = "Configured Buckets"
    color = "green"
  }
  series unencrypted {
    title = "Unencrypted Buckets"
    color = "red"
  }
  series nonversioned {
    title = "Non-Versioned Buckets"
    color = "orange"
  }
  axes {
    x {
      title = "Regions"
      labels = "auto"
    }
    y {
      title  = "Totals"
      labels = {
        display = "show"
      }
      min    = 0
      max    = 100
      steps  = 10
    }
  }
}
```

### `pie`

Display a `pie` chart. Maximum 10 categories. Automatically puts the smallest category values into an `Other` category for large category counts.

#### Data Format

# TODO single series like simple charts

| Category   | Value   |
| ---------- | ------- |
| Category 1 | Value 1 |
| Category 2 | Value 2 |
| ...        | ...     |
| Category N | Value N |

#### Properties

- `title`: See [title](#title).
- `sql`: Inline SQL or named query.
- `legend`: See [legend](#legend).
- `grouping`: `auto` (default) will automatically group the smallest values into an `Other` category when the number of categories is greater than 10, ensuring a maximum of 10 categories, or `off` will attempt to display the number of categories present.
- `categories`: A named block matching the name of the category you wish to configure. See [categories](#categories).

##### Basic

Accept all defaults.

```hcl
pie {
  sql = query.aws_a3_buckets_by_region.sql
  title = "AWS S3 Buckets by Region"
}
```

##### Kitchen Sink

Specify all options.

```hcl
pie {
  sql = query.aws_a3_unencrypted_and_nonversioned_buckets_by_region.sql
  title = "Unencrypted and Non-Versioned Buckets by Region"
  legend = {
    display  = "auto"
    position = "top"
  }
  series other {
    title = "Configured Buckets"
    color = "green"
  }
  series unencrypted {
    title = "Unencrypted Buckets"
    color = "red"
  }
  series nonversioned {
    title = "Non-Versioned Buckets"
    color = "orange"
  }
  axes {
    x {
      title = "Regions"
      labels = "auto"
    }
    y {
      title  = "Totals"
      labels = {
        display = "show"
      }
      min    = 0
      max    = 100
      steps  = 10
    }
  }
}
```

---

## Old docs

- A `container` is a new layout resource type that composes `panel` and/or `container` blocks. Supports width property:

```hcl
container s3_layout_panel {
  width = 6

  panel s3_buckets_by_region {
    type = "barchart"
    sql  = query.aws_s3_buckets_by_region.sql
  }

  panel ec2_instances_by_region {
    type = "barchart"
    sql  = query.ec2_instances_by_region.sql
  }

  container layout {
    width = 6

    panel ec2_instances_by_region {
      type = "barchart"
      sql  = query.ec2_instances_by_region.sql
    }
  }
}
```

- A `report` can contain `panel` and/or `container` blocks. (still supports `title` property):

```hcl
report my_report {
  title = "My Report"

  panel s3_buckets_by_region {
    type = "barchart"
    sql  = query.aws_s3_buckets_by_region.sql
  }

  container layout {
    width = 6

    panel ec2_instances_by_region {
      type = "barchart"
      sql  = query.ec2_instances_by_region.sql
    }
  }
}
```

- A `panel` cannot compose anything else now. It is a leaf node in the tree:

```hcl
panel s3_buckets_by_region {
  type = "barchart"
  sql  = query.aws_s3_buckets_by_region.sql
}
```

- Only `report` and `panel` can exist at the top-level of a mod.

- To re-use an existing panel, use `base`:

```hcl
panel s3_buckets_by_region {
  type = "barchart"
  sql  = query.aws_s3_buckets_by_region.sql
}

panel re_use {
  base = panel.s3_buckets_by_region
}
```

- To re-use and override a property of an existing panel, use `base` + override its property:

```hcl
panel s3_buckets_by_region {
  type = "barchart"
  sql  = query.aws_s3_buckets_by_region.sql
}

panel re_use {
  base = panel.s3_buckets_by_region
  type = "linechart"
}
```

- To re-use an existing report, use a `panel` with `base`:

```hcl
report other_report {
  panel s3_buckets_by_region {
    type = "barchart"
    sql  = query.aws_s3_buckets_by_region.sql
  }
}

report my_report {
  panel nested_report {
    base = report.other_report
  }
}
```

- A `container` cannot be re-used unlike `panel` and `report`.

- Panel `source` is now `type` and is not parse-time checked - it's just a (required) string:

```hcl
panel s3_buckets_by_region {
  type = "barchart"
  sql  = query.aws_s3_buckets_by_region.sql
}
```

- "Core" panel properties are `type` (required), `sql` (optional, named or inline query) and `width`.

- Custom/extended properties for a `panel` are declared at the top-level of the block:

```hcl
panel s3_buckets_by_region {
  type = "barchart"
  sql  = query.aws_s3_buckets_by_region.sql
  orientation = "horizontal"
}
```

- All internal lists of `container` and `panel` inside a `report` or `container` should be stored as `children` to align with `benchmark`.

## High-level design decisions

- Mods can define >=0:
  - `query`
    - A named query that can be used to obtain data
  - `panel`
    - The building block of a report - effectively a named layout component
    - A panel may render a report primitive:
      - `report`
      - `panel`
      - `markdown`
      - `table`
      - `counter`
      - `barchart` // TODO - is this just `chart` with further properties to define, or should be explicit like this?
      - `linechart` // TODO - same as above
      - `piechart` // TODO - same as above
      - tbd...
    - A panel can compose other reports and panels, allowing interesting and more complex layouts to be achieved
    - A panel can be used in isolation if required, but is typically designed to be composed with other panels into a `report`
  - `report`
    - Conceptually no different to a `panel` in that it is a named layout component that composes `panel`s and/or `report`s
    - The key differentiator is a `report` indicates that it will provide answers to specific reporting issues through composition, whereas a `panel` may only provide a subset of data that may not provide full context if used in isolation
    - A mod is expected to provide a `main` report, which is the default report to use
- All reporting primitives must know how to display themselves across a variety of screen sizes and mediums
  - The current in-scope mediums are:
    - `web`
      - Reports and panels will be mapped to a suitable grid system to provide simple, yet powerful layout constructs - spec below
      - Ideal state:
        - Support a rich and interactive reporting experience on larger devices, with graceful degradation of functionality as we move down through screen sizes
        - Panels
      - MVP state
        - Support display of reports across large and small devices e.g. columns should be full-width on small devices
    - `email`
      - Spec TBD, but likely to contain images of report primitives - pure html and attachments will ensure maximum compatibility
    - `slack`
      - Likely to have the most constrained and succinct layout options - spec TBD
- All mod component are public, meaning that anyone using a mod can import any `query`, `panel`, `report` or `action`
  - Action spec TBD - but effectively provides a consistent way to take action based on query data or the outcome of a report e.g. email a DG, send a Slack message
  - Actions are just named compositions of basic primitives including http request, send slack message, send email etc
- Namespacing:
  - Components are namespaced
    - `<org>.<component>.<primitive>`
    - e.g. `steampipe.panel.markdown`
  - Omission of namespace refers to locally-defined components within that mod e.g.
    - `panel.header`
    - `panel.body`
    - `panel.footer`
    - `query.aging_aws_iam_access_keys`
  - When building reports in custom mods you will always refer to the Steampipe reporting primitives with their full namespace e.g.
    - `steampipe.panel.markdown`
    - `steampipe.panel.counter`
    - etc...

## Primitives

Reports consist of a number of primitives, which allow various scenarios to be reported on. The following section will outline the available primitives, along with their required and optional parameters.

### Report

- A report provides a 12-column grid system. It's child panels will consume grid units from left to right, wrapping to the next row when all 12 units have been consumed, or there are insufficient grid units left on that row.
- Properties:
  - `title`: (optional) A human-friendly title for the report. This will be preferred over the HCL panel name whenever we need to present a list of reports to the user.
- Example:

```hcl
report hello_world {
  title = "Hello World"
}
```

### Panel

- A panel is the primitive that you will use to lay out your reports and present the various reporting primitives.
- Each panel will consume the full available width of its parent. E.g. if your parent panel is 12 grid units, you will have up to 12 grid unit available, however if your parent panel/report only has 6/12 grid units, your panel will only have up to 6 grid units available.
- Properties:

  - `source`: (required) The source type of this panel, which will be one of the following:

    - `steampipe.panel.markdown`: A panel for rendering markdown
    - `steampipe.panel.counter`: A panel for displaying counter data (e.g. totals)
    - `steampipe.panel.table`: A panel for displaying tabular data
    - `steampipe.panel.bar_chart`: A panel for displaying tabular data in a bar chart
    - `steampipe.panel.line_chart`: A panel for displaying tabular data in a line chart
    - `steampipe.panel.pie_chart`: A panel for displaying tabular data in a pie chart

    - `steampipe.panel.control_table`: A panel for displaying tabular data that conforms to the Steampipe [control query spec](https://steampipe.io/docs/using-steampipe/writing-controls).

  - `width`: (optional) The amount of grid unit that it should consume from its parent.
    - If not specified this will default to 12 and the panel will consume the full width of its parent.
    - Min: `1`
    - Max: `12`
  - `height`: (optional) A number of units relative to one grid unit width of itself.
    - E.g. if `height` is 4 and the panel is 1200px wide / 12 grid units wide, then 1 grid unit width is 120px, so 4 height units would be 480px.
    - If not specified the panel will be as tall as required to render its content

### Markdown

Allow rendering of markdown content in a report.

- `source: steampipe.panel.markdown`
- Supports [GitHub Flavored Markdown Spec](https://github.github.com/gfm/)
- If a panel `height` is specified and the content height exceeds the calculated height, the panel will scroll vertically.
- Properties:
  - `text`: (required) The markdown string

```hcl
panel markdown {
  source = steampipe.panel.markdown
  text = <<-EOT
    # It's just markdown
    So go `wild`.
  EOT
}
```

### Counter

Display a counter in a report.

- `steampipe.panel.counter`
- Expects data in format

| \<Title of counter\> |
| -------------------- |
| \<Counter value\>    |

e.g.

| AWS IAM Users |
| ------------- |
| 17            |

- First row is expected to be the header with title
- Second row is the counter value itself

```hcl
panel aws_iam_user_count {
  source = steampipe.panel.counter
  query = "select count(*) as "AWS IAM Users" from aws_iam_user"
}
```

### Table

Allow rendering of table data in a report.

- `source: steampipe.panel.table`
- Expects data in table (e.g. like CSV/Excel) format. First row assumed to be the header row. Subsequent rows assumed to be the data rows.
- Properties:
  - `sql`: Either an inline query, or a named query

Example with inline query:

```hcl
panel buckets_versioning_disabled {
  source = steampipe.panel.table
  sql = "select name as Name, region as Region from aws_s3_bucket where versioning = 'disabled'"
}
```

Example with named query:

```hcl
panel buckets_versioning_disabled {
  source = steampipe.panel.table
  sql = query.buckets_versioning_disabled_query.sql
}
```

### Pie Chart

Allow rendering of data as a pie chart in a report.

- `steampipe.panel.piechart`
- First column will be projected across the X-axis
- Second column will be used as the series data
- Expects data in table format

| Entity   | Total |
| -------- | ----- |
| Groups   | 2     |
| Policies | 102   |
| ...      | ...   |
| Users    | 10    |

```hcl
panel buckets_versioning_disabled {
  source = steampipe.panel.barchart
  title = "AWS IAM Entities"
  query = "select entity as Entity, total as Total;"
}
```

### Bar Chart

Allow rendering of data as a bar chart in a report.

- `steampipe.panel.barchart`
- First column will be projected across the X-axis
- Second column will be used as the series data
- Expects data in table format

| Entity   | Total |
| -------- | ----- |
| Groups   | 2     |
| Policies | 102   |
| ...      | ...   |
| Users    | 10    |

```hcl
panel buckets_versioning_disabled {
  source = steampipe.panel.barchart
  title = "AWS IAM Entities"
  query = "select entity as Entity, total as Total;"
}
```

### Stacked Bar Chart

Allow rendering of data as a stacked bar chart in a report.

- `steampipe.panel.stackedbarchart`
- First column will be projected across the X-axis
- Each subsequent column will be used as the series data to stack
- Expects data in table format

| Entity   | Used | Available |
| -------- | ---- | --------- |
| Groups   | 2    | 498       |
| Policies | 102  | 898       |
| ...      | ...  | ...       |
| Users    | 10   | 90        |

```hcl
panel buckets_versioning_disabled {
  source = steampipe.panel.stackedbarchart
  title = "AWS IAM Entity Quota Usage"
  query = "select entity as Entity, total as Used, (total - quota) as Available from aws_iam_account_summary;"
}
```

### Line Chart

Allow rendering of data as a line chart in a report.

- `steampipe.panel.linechart`
- First column will be projected across the X-axis
- Second column will be used as the series data
- Expects data in table format

#### Single Line

| Entity   | Total |
| -------- | ----- |
| Groups   | 2     |
| Policies | 102   |
| ...      | ...   |
| Users    | 10    |

```hcl
panel buckets_versioning_disabled {
  source = steampipe.panel.linechart
  title = "AWS IAM Entities"
  query = "select entity as Entity, total as Total;"
}
```

#### Multi Line

| Entity   | Used | Available |
| -------- | ---- | --------- |
| Groups   | 2    | 498       |
| Policies | 102  | 898       |
| ...      | ...  | ...       |
| Users    | 10   | 90        |

```hcl
panel buckets_versioning_disabled {
  source = steampipe.panel.linechart
  title = "AWS IAM Entity Quota Usage"
  query = "select entity as Entity, total as Used, (total - quota) as Available from aws_iam_account_summary;"
}
```

## Examples

### Report with single markdown panel

- Single full-width panel rendered with a title in markdown
  - <!---TODO: text? check this-->

```hcl
report hello_world {

  panel header {
    source = steampipe.panel.markdown
    text = "# Hello World!"
  }

}
```

### 2-column, 1-row layout, single panel in each column

- Top-level panel, but specifying a width of 6 columns each

```hcl
report two_column_one_row {

  panel left {
    width = 6
    source = steampipe.panel.markdown
    text = "# Left Side"
  }

  panel right {
    width = 6
    source = steampipe.panel.markdown
    text = "# Right Side"
  }

}
```

### 1-column, 2-row layout, single panel in each row

- Top-level panel, so implicitly 12 columns wide
- Child panels
- No height specified, so will fill available height

```hcl
report one_column_two_row {

  panel main {

    panel first {
      source = steampipe.panel.markdown
      text = "# First Row"
    }

    panel second {
      source = steampipe.panel.markdown
      text = "# Second Row"
    }

  }

}
```

### Basic report with header, nav, content and footer

```hcl
report basic {

  panel header {
    source = steampipe.panel.markdown
    text = "# Header"
  }

  panel main {

    panel nav {
      width = 3
      source = steampipe.panel.markdown
      text = "# Nav"
    }

    panel content {
      width = 9
      source = steampipe.panel.markdown
      text = "# Content"
    }

  }

  panel footer {
    source = steampipe.panel.markdown
    text = "# Footer"
  }

}
```

### 2-column, 1-row layout with nested panels

- Top-level panel, so implicitly 24 columns wide
- No height specified, so will fill available height

```hcl
panel header_top_row {

  panel left {
    width = 6
    source = steampipe.panel.markdown
    text = "# Left"
  }

  panel right {
    width = 6
    source = steampipe.panel.markdown
    text = "# Right"
  }

}

panel main {

  panel header {
    source = panel.header_top_row
  }

}
```

### 2-column, 1-row layout with 2 nested panels each containing 2 columns

- Top-level panel, so implicitly 24 columns wide
- No height specified, so will fill available height

```hcl
panel content_left {

  panel left {
    width = 6
    source = steampipe.panel.markdown
    text = "# One"
  }

  panel right {
    width = 6
    source = steampipe.panel.markdown
    text = "# Two"
  }

}

panel content_right {

  panel left {
    width = 6
    source = steampipe.panel.markdown
    text = "# Three"
  }

  panel right {
    width = 6
    source = steampipe.panel.markdown
    text = "# Four"
  }

}

panel main {

  panel left {
    width = 6
    source = panel.content_left
  }

  panel right {
    width = 6
    source = panel.content_right
  }

}
```

## Primitives

### Panel

The basic layout construct for reports

```hcl
panel header {
  source = steampipe.panel.markdown
  text = "# Render me"
}
```

### Example AWS Report

This example takes a few of the report primitives outlined above and composes them into a simple report

```hcl
report aws {

  panel header {
    source = steampipe.panel.markdown
    text = <<-EOT
      # AWS Report
      Overview of a variety of key AWS metrics.
    EOT
  }

  panel iam {

    panel iam-header {
      source = steampipe.panel.markdown
      text = "### IAM"
    }

    panel iam-count {
      width = 3
      source = steampipe.panel.counter
      query = "select count(*) as "Unused Access Keys" from aws_iam_user"
    }

    panel iam-entities {
      width = 9
      source = steampipe.panel.barchart
      query = "select groups, groups_quota, policies, policies_quota, roles, roles_quota, users, users_quota from aws_iam_account_summary;"
    }

  }

  panel cost {
    width = 8

    panel cost-header {
      source = steampipe.panel.markdown
      text = "### Cost"
    }

    panel cost-monthly {
      source = steampipe.panel.linechart
      query = "select period_start as "Period Start", unblended_cost_amount as "Unblended Cost" from aws_cost_by_account where granularity = 'monthly' order by period_start;"
    }

  }

  panel lambda {
    width = 4

    panel lambda-header {
      source = steampipe.panel.markdown
      text = "### Lambda"
    }

    panel lambda-runtimes {
      source = steampipe.panel.piechart
      query = "select runtime as "Runtime", count(runtime) as "Count" from aws_lambda_function group by runtime order by runtime;"
    }

  }

  panel kms {

    panel kms-header {
      source = steampipe.panel.markdown
      text = "### KMS"
    }

    panel kms-keys {
      source = steampipe.panel.table
      query = "select title as "Id", region as "Region", key_rotation_enabled as "Rotation Enabled", description as "Description" from aws_kms_key;"
    }

  }

}
```
