
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <meta name="description" content="">
    <meta name="author" content="">

    <title>Prometheus Metric Analyzer</title>

    <!-- Custom styles for this template -->
    <link href="navbar-top-fixed.css" rel="stylesheet">

    <!-- Latest compiled and minified CSS -->
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@3.3.7/dist/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">

    <!-- Optional theme -->
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@3.3.7/dist/css/bootstrap-theme.min.css" integrity="sha384-rHyoN1iRsVXV4nD0JutlnGaslCJuC7uwjduW9SVrLvRYooPp2bWYgmgJQIXwl/Sp" crossorigin="anonymous">
 
    <style  rel="stylesheet" type="text/css">
      table, th, td {
        border: 1px solid;
        padding-left: 1em;
        padding-right: 1em;
        text-align: center;
      }
    </style>


    <script src="https://code.jquery.com/jquery-3.6.0.min.js"></script>

    
  </head>

  <body>


   <div class="container-summary" style="padding: 2em;">
    <h1>Notice</h1>
    <p class="lead" style="text-align: left; font-size: 1.4em;">
      Total linting warnings: <strong>{{ .totalLintingProblems }}</strong><br/>
      Total metrics: <strong>{{ .totalMetrics }}</strong>
    </p><br/>
   </div>

   <hr/>

    <div class="container">

      <div class="starter-template">

        <div class="headerLinting"><h1>Linting Stats</h1> (click to show/hide)</div>
        <div class="hideLinting" style="display:none" >
          
          <p class="lead">
            {{ if .lintingProblems }}
            <table class="table-results">
              <tr>
                <th>Metric Name</th>
                <th>Problem</th>
              </tr>
            {{ range $key, $p := .lintingProblems }}
              <tr>
                <td><strong>{{ $p.Metric }}</strong></td>
                <td style="font-size: 1.2em;">{{ $p.Text }}</td>
              </tr>
            {{ end }}
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
              <tr>
                <th>Metric Name</th>
                <th>Cardinality</th>
                <th>Total Percentage</th>
              </tr>
            {{ range $key, $value := .resultCardinality }}
              <tr>
                <td><strong>{{ $value.Name }}</strong></td>
                <td>{{ $value.Cardinality }}</td>
                <td>{{ printf "%.2f" $value.Percentage }}</td>
              </tr>
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
      </div>

    </div><!-- /.container -->

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
