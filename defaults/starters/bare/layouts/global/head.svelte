<script>
  export let title, env;
</script>

<head>
  <meta charset='utf-8'>
  <meta name='viewport' content='width=device-width,initial-scale=1'>

  <title>{ title }</title>

  <script type="module" src={`/${env.entrypointJS}/core/main.js`}></script>

  <link rel="icon" type="image/svg+xml" href="/media/logo.svg">
  <link rel='stylesheet' href={`/${env.entrypointJS}/bundle.css`}>
</head>
