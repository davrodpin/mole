{{.mole.Title}}

Current Commit : {{.mole.Current_Commit}}
Previous Commit: {{.mole.Previous_Commit}}
{{ range $file := .mole.Files}}
File: {{$file.File}}
  Previous Coverage: {{$file.Previous_Coverage}}%
  Current Coverage : {{$file.Current_Coverage}}%
  Diff             : {{$file.Diff}}%
{{ end}}

