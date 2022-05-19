
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
        background-color: blue;
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

  <h1>Prometheus Metrics Endpoint Analyzer</h1>

  <form action="/analyze" method="post" enctype="application/x-www-form-urlencoded">
      URL: <input type="text" name="url"><br>
      <input type="submit" value="Submit">
  </form>

    <!-- Bootstrap core JavaScript
    ================================================== -->
    <!-- Placed at the end of the document so the pages load faster -->
    <script src="https://ajax.googleapis.com/ajax/libs/jquery/1.12.4/jquery.min.js"></script>
    <script>window.jQuery || document.write('<script src="../../assets/js/vendor/jquery.min.js"><\/script>')</script>
    <!-- Latest compiled and minified JavaScript -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@3.3.7/dist/js/bootstrap.min.js" integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa" crossorigin="anonymous"></script>
  </body>
</html>
