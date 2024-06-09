import { useState } from "react";
import { ToolBar } from "./components/ToolBar";
import { Log } from "./components/Log";

function App() {
  const [connected, setConnected] = useState(false);
  return (
    <div
      style={{
        display: "flex",
        gap: "28px",
      }}
    >
      <ToolBar connected={connected} setConnected={setConnected} />
      <Log connected={connected} />
    </div>
  );
}

export default App;
