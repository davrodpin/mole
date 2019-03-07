<html>
  <head>
    <title>{{.mole.Title}}</title>
    <meta charset="UTF-8">
    <meta name="description" content="Free Web tutorials">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/mini.css/3.0.1/mini-default.min.css">
  </head>
  <body>

    <header>
      <a href="https://davrodpin.github.io/mole/" class="logo">Mole - Easily create SSH Tunnels</a>
    </header>

    <div class="container">
      <div class="row">
        <div class="col-lg">
          <h3>Code coverage comparison between <a href="https://github.com/davrodpin/mole/commit/{{.mole.Previous_Commit}}">{{.mole.Previous_Commit}}</a> and <a href="https://github.com/davrodpin/mole/commit/{{.mole.Current_Commit}}">{{.mole.Current_Commit}}</a> </small></h3>
        </div>
      </div>
      <div class="row">
        <div class="col-lg">
          <table class="hoverable">
            <thead>
              <tr>
                <th>File</th>
                <th>Previous Coverage (<a href="https://github.com/davrodpin/mole/commit/{{.mole.Previous_Commit}}">{{.mole.Previous_Commit}}</a>)</th>
                <th>Current Coverage (<a href="https://github.com/davrodpin/mole/commit/{{.mole.Current_Commit}}">{{.mole.Current_Commit}}</a>)</th>
                <th>Delta</th>
              </tr>
            </thead>
            <tbody>
              {{- range $file := .mole.Files}}
                <tr>
                  <td>{{$file.File}}</td>
                  <td>{{printf "%.2f" $file.Previous_Coverage}}%</td>
                  <td>{{printf "%.2f" $file.Current_Coverage}}%</td>
                  {{- if (eq $file.Diff 0.0)}}
                  <td>{{$file.Diff}}%</td>
                  {{- else if (gt $file.Diff 0.0)}}
                  <td><mark class="tertiary">+{{$file.Diff}}%</mark></td>
                  {{- else}}
                  <td><mark class="secondary">{{$file.Diff}}%</mark></td>
                  {{- end}}
                </tr>
              {{- end}}
            </tbody>
          </table>
        </div>
      </div>
    </div>

  </body>
</html>
