package defaults

// Public : this is a TEMPORARY defualt for the build folder.
// Once the HTML is created automatically this should be removed.
var Public = map[string][]byte{
	"/public/index.html": []byte(`<!DOCTYPE html>
	<html lang="en" id="hydrate-plenti">
	<head>
		<meta charset='utf-8'>
		<meta name='viewport' content='width=device-width,initial-scale=1'>
	
		<title>Plenti app</title>
	
		<link rel='icon' type='image/png' href='/favicon.png'>
		<link rel='stylesheet' href='/build/bundle.css'>
	</head>
	
	<body>
		<main>
		<h1>no js</h1>
		</main>
		<script type="module" src="https://unpkg.com/dimport?module" data-main="/build/spa/main.js"></script>
		<script nomodule src="https://unpkg.com/dimport/nomodule" data-main="/build/spa/main.js"></script>
	</body>
	
	
	</html>`),
}
