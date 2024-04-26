<script>
  export let title, env;
</script>

<head>
  <meta charset='utf-8'>
  <meta name='viewport' content='width=device-width,initial-scale=1'>

  <title>{ title }</title>

  <script type="module" src="/spa/core/main.js?{env.fingerprint}"></script>

  <link rel="icon" type="image/svg+xml" href="/media/logo.svg">
  <link rel='stylesheet' href="/spa/bundle.css?{env.fingerprint}">
</head>
