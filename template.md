# Contributors of {{.Org}}/{{.Repo}}

{{range $c := .Contributors}}
[![{{$c.Name}}](avatars/{{$c.Name}}.png)]({{$c.ProfileURL}})
{{end}}
