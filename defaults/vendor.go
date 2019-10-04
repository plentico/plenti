package defaults

var Vendor = map[string][]byte{
	"/.gitignore": []byte(`public
node_modules`),
	"/package.json": []byte(`{
  "name": "plenti-default",
  "version": "1.0.0",
  "description": "",
  "main": "index.js",
  "scripts": {
    "test": "echo \"Error: no test specified\" && exit 1"
  },
  "author": "",
  "license": "ISC",
  "dependencies": {
    "react": "^16.10.1",
		"react-dom": "^16.10.2"
	},
	"devDependencies": {
		"@babel/preset-react": "^7.0.0"
	}
}`),
	"/.babelrc": []byte(`{
  "presets": [
    "@babel/preset-env",
    "@babel/preset-react"
  ]
}`),
	"/webpack.config.js": []byte(`module.exports = {
  entry: './src/index.js',
  module: {
    rules: [
      {
        test: /\.(js|jsx)$/,
        exclude: /node_modules/,
        use: ['babel-loader']
      }
    ]
  },
  resolve: {
    extensions: ['*', '.js', '.jsx']
  },
  output: {
    path: __dirname + '/dist',
    publicPath: '/',
    filename: 'bundle.js'
  },
  devServer: {
    contentBase: './dist'
  }
};`),
}
