{{- define "yardpass-bot.release-name" -}}
{{ .Release.Name }}
{{- end }}

{{- define "yardpass-bot.resource-name" -}}
{{- .Values.appLabel -}}
{{- end }}

{{- define "yardpass-bot.selector" -}}
app: {{ .Values.appLabel }}
{{- end }}

{{- define "yardpass-bot.labels" -}}
app: {{ .Values.appLabel }}
{{- end }}
