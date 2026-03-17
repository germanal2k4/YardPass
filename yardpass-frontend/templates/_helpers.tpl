{{- define "yardpass-frontend.release-name" -}}
{{ .Release.Name }}
{{- end }}

{{- define "yardpass-frontend.resource-name" -}}
{{- .Values.appLabel -}}
{{- end }}

{{- define "yardpass-frontend.selector" -}}
app: {{ .Values.appLabel }}
{{- end }}

{{- define "yardpass-frontend.labels" -}}
app: {{ .Values.appLabel }}
{{- end }}

{{- define "yardpass-frontend.liveness" -}}
{{- if .Values.livenessProbe.enabled }}
livenessProbe:
  httpGet:
    path: {{ .Values.livenessProbe.httpGet.path }}
    port: {{ .Values.livenessProbe.httpGet.port }}
  initialDelaySeconds: {{ .Values.livenessProbe.initialDelaySeconds }}
  periodSeconds: {{ .Values.livenessProbe.periodSeconds }}
{{- end }}
{{- end }}

{{- define "yardpass-frontend.readiness" -}}
{{- if .Values.readinessProbe.enabled }}
readinessProbe:
  httpGet:
    path: {{ .Values.readinessProbe.httpGet.path }}
    port: {{ .Values.readinessProbe.httpGet.port }}
  initialDelaySeconds: {{ .Values.readinessProbe.initialDelaySeconds }}
  periodSeconds: {{ .Values.readinessProbe.periodSeconds }}
{{- end }}
{{- end }}
