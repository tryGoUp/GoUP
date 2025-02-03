const express = require("express");

const app = express();
const PORT = process.env.PORT || 3000;

app.use((req, res, next) => {
  console.log(`[NodeJSPlugin] Request: ${req.method} ${req.url}`);
  next();
});

app.get("/api/test", (req, res) => {
  res.json({ message: "Hello from NodeJSPlugin!", timestamp: Date.now() });
});

app.listen(PORT, () => {
  console.log(`NodeJS server running on http://localhost:${PORT}`);
});
