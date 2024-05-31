import { useState } from "react";

function App() {
  return (
    <>
      <ToolBar />
      <Log />
    </>
  );
}

function ToolBar() {
  const [ip, setIp] = useState("");
  const [port, setPort] = useState("");

  const handleSubmit = () => {
    alert(`IP Address: ${ip}, Port: ${port}`);
  };

  return (
    <div>
      <div>
        <label>
          IP Address:
          <input
            type="text"
            value={ip}
            onChange={(e) => setIp(e.target.value)}
            placeholder="Enter IP address"
          />
        </label>
      </div>
      <div>
        <label>
          Port:
          <input
            type="text"
            value={port}
            onChange={(e) => setPort(e.target.value)}
            placeholder="Enter port"
          />
        </label>
      </div>
      <button onClick={handleSubmit}>Submit</button>
    </div>
  );
}

function Log() {
  return <></>;
}

export default App;
