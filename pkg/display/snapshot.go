package display

const snapshotTemplate = `
{
  "schema_version": "{{ .SchemaVersion }}",
  "panels": {
    {{- range $panel := .Panels }}
    "{{ $panel.Dashboard }}": {
      "dashboard": "{{ $panel.Dashboard }}",
      "name": "{{ $panel.Name }}",
      "panel_type": "{{ $panel.PanelType }}",
      "source_definition": "{{ $panel.SourceDefinition }}",
      "status": "{{ $panel.Status }}",
      "title": "{{ $panel.Title }}"
    },
    "custom.table.results": {
      "dashboard": "{{ $panel.Dashboard }}",
      "name": "custom.table.results",
      "panel_type": "table",
      "source_definition": "{{ $panel.SourceDefinition }}",
      "status": "{{ $panel.Status }}",
      "sql": "{{ $panel.SQL }}",
      "properties": {
        "name": "results"
      },
      {{- if $panel.Data }}
      "data": {
        "columns": [
          {{- range $i, $col := $panel.Data.Columns }}
          {
            "name": "{{ $col.Name }}",
            "data_type": "{{ $col.DataType }}",
            "original_name": "{{ $col.OriginalName }}"
          }{{ if lt (add1 $i) (len $panel.Data.Columns) }},{{ end }}
          {{- end }}
        ],
        "rows": [
          {{- range $rowIndex, $row := $panel.Data.Rows }}
          {
            {{- $rowLen := len $row }}
            {{- $currentIndex := 0 }}
            {{- range $key, $value := $row }}
            "{{ $key }}": {{ $value }}{{ if lt (add1 $currentIndex) $rowLen }},{{ end }}
            {{- $currentIndex = add1 $currentIndex }}
            {{- end }}
          }{{ if lt (add1 $rowIndex) (len $panel.Data.Rows) }},{{ end }}
          {{- end }}
        ]
      }
      {{- end }}
    }
    {{- end }}
  },
  "inputs": {},
  "variables": {},
  "search_path": [
    {{- range $i, $path := .SearchPath }}
    "{{ $path }}"{{ if lt (add1 $i) (len $.SearchPath) }},{{ end }}
    {{- end }}
  ],
  "start_time": "{{ .StartTime }}",
  "end_time": "{{ .EndTime }}",
  "layout": {
    "name": "{{ .Layout.Name }}",
    "children": [
      {
        "name": "custom.table.results",
        "panel_type": "table"
      }
    ],
    "panel_type": "dashboard"
  }
}
`
