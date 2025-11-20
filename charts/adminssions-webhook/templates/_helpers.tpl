{{/*
Return the fully qualified name of the chart
*/}}
{{- define "admissions-webhook.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}