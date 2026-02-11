terraform {
  required_version = ">= 1.2.0"

  required_providers {
{{- range .Providers }}
    {{ .Name }} = {
      source  = "{{ .Source }}"
      version = "{{ .Version }}"
    }
{{- end}}
  }
}

{{- range $key := .Providers }}

provider "{{ $key.Name }}" {
{{- if eq $key.Name "azurerm" }}
  features {}
{{- end }}
}
{{- end }}

module "m" {
  source = "{{ .Module.Source }}"
{{- if .Module.Version }}
  version = "{{ .Module.Version }}"
{{- end }}

{{- range $key, $value := .Module.Variables }}
  {{- if needsQuotes $value }}
  {{ $key }} = "{{ $value }}"
  {{- else }}
  {{ $key }} = {{ $value }}
  {{- end }}
{{- end }}
}

output "values" {
  value = module.m
}
