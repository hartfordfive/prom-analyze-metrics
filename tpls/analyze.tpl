
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <meta name="description" content="">
    <meta name="author" content="">

    <title>Prometheus Metric Analyzer</title>

    <!-- Latest compiled and minified CSS -->
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@3.3.7/dist/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">

    <!-- Optional theme -->
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@3.3.7/dist/css/bootstrap-theme.min.css" integrity="sha384-rHyoN1iRsVXV4nD0JutlnGaslCJuC7uwjduW9SVrLvRYooPp2bWYgmgJQIXwl/Sp" crossorigin="anonymous">
 
    <style  rel="stylesheet" type="text/css">
      table th {
        padding-left: 1em;
        padding-right: 1em;
        text-align: center;
      }

      table {
        border-collapse: collapse;
          font-family: Tahoma, Geneva, sans-serif;
      }
      table td {
        padding: 15px;
      }
      table thead td {
        background-color: #ECEFF0;
        color: #000000;
        font-weight: bold;
        font-size: 13px;
        text-align: center;
        border: 1px solid #54585d;
      }
      table tbody td {
        color: #636363;
        border: 1px solid #dddfe1;
      }
      table tbody tr {
        background-color: #f9fafb;
      }
      table tbody tr:nth-child(odd) {
        background-color: #ffffff;
      }

      input[type=text], select {
        width: 100%;
        padding: 12px 20px;
        margin: 8px 0;
        display: inline-block;
        border: 1px solid #ccc;
        border-radius: 4px;
        box-sizing: border-box;
      }

      input[type=submit] {
        width: 100%;
        background-color: #4CAF50;
        color: white;
        padding: 14px 20px;
        margin: 8px 0;
        border: none;
        border-radius: 4px;
        cursor: pointer;
      }

      input[type=submit]:hover {
        background-color: #45a049;
      }
    </style>


    <script src="https://code.jquery.com/jquery-3.6.0.min.js"></script>

    
  </head>

  <body>

   <div class="container-summary" style="padding: 2em;">

    <h1>Prometheus Metrics Endpoint Summary</h1>
    <p class="lead" style="text-align: left; font-size: 1.2em;">
        <table>
          {{- if .url }}
          <tr>
            <td>Target</td>
            <td><strong>{{ .url }}</strong></td>
          </tr>
          {{- end }}
          {{- if .transferSize }}
          <tr>
            <td>Transfer Size</td>
            <td><strong>{{ bytesToHuman .transferSize }}</strong></td>
          </tr>
          {{- end }}
          {{- if .totalLintingProblems }}
          <tr>
            <td>Total linting warnings</td>
            <td><strong>{{ .totalLintingProblems }}</strong></td>
          </tr>
          {{- end }}
          {{- if .totalMetrics }}
          <tr>
            <td>Total metrics</td>
            <td><strong>{{ .totalMetrics }}</strong></td>
          </tr>
          {{- end }}
          {{- if .resultCardinality }}
          <tr>
            {{ $topMetric := index .resultCardinality 0 }}
            <td>Highest cardinality metric</td>
            <td><strong><strong>{{ $topMetric.Name }} ({{ $topMetric.Cardinality }} total combinations)</strong></strong></td>
          </tr>
          {{- end }}
          {{- if .error }}
          <tr>
            <td>Errors</td>
            <td>
            {{ .error }}<br/>
            </td>
          </tr>
          {{- end }}
          </table>
    </p><br/>
    <a href="/">Check Another Endpoint</a>
   </div>

   <hr/>

    {{- if not .error }}
    <div class="container">

  
        <div class="headerLinting"><h1>Linting Stats</h1> (click to show/hide)</div>
        <div class="hideLinting" style="display:none" >
          
          <p class="lead">
            {{ if .lintingProblems }}
            <table class="table-results">
              <thead>
                <tr>
                  <td>Metric Name</td>
                  <td>Problem</td>
                </tr>
              <thead>
              <tbody>
              {{ range $key, $p := .lintingProblems }}
                <tr>
                  <td><strong>{{ $p.Metric }}</strong></td>
                  <td style="font-size: 1.2em;">{{ $p.Text }}</td>
                </tr>
              {{ end }}
              <tbody>
            </table>
            {{ else }}
              No linting problems found.
            {{ end }}
          </p>


        </div> <!-- END OF DIF content -->
        <br/><br/>

        <div class="headerCardinality"><h1>Cardinality Stats</h1> (click to show/hide)</div>
        <div class="hideCardinality" style="display:none" >

          <p class="lead">
            <table class="table-results">
              <thead>
                <tr>
                  <td>Metric Name</td>
                  <td>Cardinality</td>
                  <td>Total Percentage</td>
                </tr>
              </thead>
            {{ range $key, $value := .resultCardinality }}
              <tbody>
                <tr>
                  <td><strong>{{ $value.Name }}</strong></td>
                  <td>{{ $value.Cardinality }}</td>
                  {{ $percentage := floatToPercentage $value.Percentage }}
                  {{ if eq $percentage 0.00}}
                  <td>< 1%</td>
                  {{ else }}
                  <td>{{ $percentage }}%</td>
                  {{ end }}
                </tr>
              </tbody>
            {{ end }}
            <tr>
                <td colspan="3" style="padding: 1em;"></td>
            </tr>
            <tr>
                <td style="text-align: right; font-size: 1.4em;">Total Metrics</td>
                <td style="text-align: center; font-size: 1.4em;"><strong>{{ .totalMetrics }}</strong></td>
                <td></td>
            </tr>
            </table>
          </p>
        </div>
    </div><!-- /.container -->
    {{- end }}

    <script>
      $('.headerLinting').click(function(){
          $('.hideLinting').toggle();
      });
      $('.headerCardinality').click(function(){
          $('.hideCardinality').toggle();
      });
    </script>

    <!-- Bootstrap core JavaScript
    ================================================== -->
    <!-- Placed at the end of the document so the pages load faster -->
    <script src="https://ajax.googleapis.com/ajax/libs/jquery/1.12.4/jquery.min.js"></script>
    <script>window.jQuery || document.write('<script src="../../assets/js/vendor/jquery.min.js"><\/script>')</script>
    <!-- Latest compiled and minified JavaScript -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@3.3.7/dist/js/bootstrap.min.js" integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa" crossorigin="anonymous"></script>
  </body>
</html>
