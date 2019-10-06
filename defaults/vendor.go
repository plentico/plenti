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
    "test": "echo \"Error: no test specified\" && exit 1",
    "dev": "webpack --mode development"
  },
  "author": "",
  "license": "ISC",
  "dependencies": {
    "@babel/core": "^7.6.2",
    "@babel/preset-env": "^7.6.2",
    "react": "^16.10.1",
    "react-dom": "^16.10.2"
  },
  "devDependencies": {
    "@babel/preset-react": "^7.0.0",
    "babel-loader": "^8.0.6",
    "webpack": "^4.41.0",
    "webpack-cli": "^3.3.9"
  }
}`),
	"/.babelrc": []byte(`{
  "presets": [
    "@babel/preset-env",
    "@babel/preset-react"
  ]
}`),
	"/webpack.config.js": []byte(`module.exports = {
  entry: './templates/entry.js',
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
    path: __dirname + '/public/dist',
    publicPath: '/',
    filename: 'bundle.js'
  },
  devServer: {
    contentBase: './public/dist'
  }
};`),
}
