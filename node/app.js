const express = require('express');
const app = express();

const port = process.env.PORT;

if (!port) {
  console.error('Error: The environment variable PORT is not set.');
  process.exit(1);
}

app.get('/', (req, res) => {
  res.send('Hello, World from Node.js!');
});

app.listen(port, '0.0.0.0', () => {
  console.log(`Node.js app listening at http://0.0.0.0:${port}`);
});
