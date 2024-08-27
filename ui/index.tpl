<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="ie=edge">
  <title>Shorty</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@1/css/pico.min.css">
  <script src="https://cdn.jsdelivr.net/npm/htmx.org@1/dist/htmx.min.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/htmx.org@1/dist/ext/ws.min.js"></script>
</head>

<body hx-ext="ws" ws-connect="/ws">
  <nav class="container-fluid">
    <ul>
      <li><a href="/logout">Logout</a></li>
    </ul>
  </nav>
  <main class="container-fluid">
    <table role="grid">
      <thead>
        <th>Shorty</th>
        <th>File</th>
        <th>Url</th>
        <th>Expired</th>
      </thead>
      <tbody id="tbody"></tbody>
    </table>
</body>

</html>
