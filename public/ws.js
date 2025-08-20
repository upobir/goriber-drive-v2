import { state, renderFiles } from "./render.js";

const setupSocket = () => {
  const port = location.port ? `:${location.port}` : "";
  const ws = new WebSocket(`ws://${location.hostname}${port}/ws`);

  const wsHandlers = {
    ["welcome"]: (data, timestamp) => {
      console.log(`Welcome message: ${data}, at timestamp: ${timestamp}`);
    },
    ["file.created"]: (data, timestamp) => {
      state.files.push(data);
      renderFiles();
    },
    ["file.deleted"]: (data, timestamp) => {
      state.files = state.files.filter((file) => file.id !== data);
      renderFiles();
    },
  };

  ws.onopen = () => console.log("ws: open");
  ws.onmessage = (e) => {
    console.log("ws: message received", e.data);
    try {
      const event = JSON.parse(e.data);
      const handler = wsHandlers[event.type];
      if (handler) {
        handler(event.data, event.timestamp);
      }
    } catch (error) {
      console.error("Error parsing event:", error);
    }
  };
  ws.onclose = () => console.log("ws: closed");
};

export { setupSocket };
