<script>
  export let title, env;
</script>

<head>
  <meta charset='utf-8'>
  <meta name='viewport' content='width=device-width,initial-scale=1'>

  <title>{ title }</title>

  <base href="{ env.local ? '/' : env.baseurl }">
  <script type="module" src="spa/ejected/main.js"></script>

  <link href="https://fonts.googleapis.com/css2?family=Rubik:ital,wght@0,300;0,700;1,300&display=swap" rel="stylesheet">
  <link rel="icon" type="image/svg+xml" href="assets/logo.svg">
  <link rel='stylesheet' href='spa/bundle.css'>
</head>
