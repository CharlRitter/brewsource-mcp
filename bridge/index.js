// bridge.js
// MCP WebSocket proxy to dev (Tilt/kind) or prod endpoints based on MODE env
// Usage: configured as a stdio MCP server

const WebSocket = require('ws');
process.stdin.setEncoding('utf8');

const MODE = process.env.MODE || 'dev';
const DEV_WS_URL = 'ws://localhost:8080/mcp';
const PROD_WS_URL = 'wss://prod.example.com/mcp';
const WS_URL = MODE === 'prod' ? PROD_WS_URL : DEV_WS_URL;

let ws;
function connectWebSocket() {
  ws = new WebSocket(WS_URL);

  ws.on('open', () => {
    process.stdin.on('data', data => {
      let reqJson;
      try {
        reqJson = JSON.parse(data);
      } catch (e) {
        process.stdout.write(JSON.stringify({ error: 'Invalid JSON' }) + '\n');
        return;
      }
      ws.send(JSON.stringify(reqJson));
    });
  });

  ws.on('message', msg => {
    process.stdout.write(msg + '\n');
  });

  ws.on('error', err => {
    process.stdout.write(JSON.stringify({ error: err.message }) + '\n');
    process.exit(1);
  });

  ws.on('close', () => {
    process.exit(0);
  });
}

connectWebSocket();
