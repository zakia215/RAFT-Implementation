import { useState } from "react";
import { connect, ping, get, set, strln, append } from "../lib/lib";

type ToolBarProps = {
  connected: boolean;
  setConnected: React.Dispatch<React.SetStateAction<boolean>>;
};

export function ToolBar({ connected, setConnected }: ToolBarProps) {
  const [ip, setIp] = useState("");
  const [port, setPort] = useState("");
  const [selectedCommand, setSelectedCommand] = useState("");
  const [key, setKey] = useState("");
  const [value, setValue] = useState("");
  const [reply, setReply] = useState("");

  const commandOptions = ["ping", "get", "set", "strln", "append"];

  const handleConnect = () => {
    connect(ip, port).then((success) => {
      setConnected(success);
    });
  };

  const handleExecute = async () => {
    let commandReply = "";
    switch (selectedCommand) {
      case "ping":
        commandReply = await ping();
        break;
      case "get":
        commandReply = await get(key);
        break;
      case "set":
        commandReply = await set(key, value);
        break;
      case "strln":
        commandReply = await strln(key);
        break;
      case "append":
        commandReply = await append(key, value);
        break;
      default:
        console.error("Invalid command");
    }
    setReply(commandReply);
  };

  const labelStyle = {
    display: "block",
  };

  const inputStyle: React.CSSProperties = {
    padding: "5px",
    fontSize: "1em",
    margin: "10px 0 10px 0",
  };

  return (
    <div style={{ paddingLeft: "40px" }}>
      <div>
        <label style={labelStyle}>IP Address:</label>
        <input
          type="text"
          value={ip}
          onChange={(e) => setIp(e.target.value)}
          placeholder="Enter IP address"
          style={inputStyle}
        />
      </div>
      <div>
        <label style={labelStyle}>Port:</label>
        <input
          type="text"
          value={port}
          onChange={(e) => setPort(e.target.value)}
          placeholder="Enter port"
          style={inputStyle}
        />
      </div>
      <button onClick={handleConnect} style={{ marginBottom: "10px" }}>
        Connect
      </button>
      <div
        style={{
          display: "flex",
          gap: "40px",
        }}
      >
        <div>
          <div>
            <label style={labelStyle}>Select a command:</label>
            <select
              value={selectedCommand}
              onChange={(e) => setSelectedCommand(e.target.value)}
              style={inputStyle}
            >
              <option value="">Choose a command</option>
              {commandOptions.map((choice, index) => (
                <option key={index} value={choice}>
                  {choice}
                </option>
              ))}
            </select>
          </div>
          {selectedCommand && selectedCommand !== "ping" && (
            <div>
              <label style={labelStyle}>Key:</label>
              <input
                type="text"
                value={key}
                onChange={(e) => setKey(e.target.value)}
                placeholder="Enter key"
                style={inputStyle}
              />
            </div>
          )}
          {(selectedCommand === "set" || selectedCommand === "append") && (
            <div>
              <label style={labelStyle}>Value:</label>
              <input
                type="text"
                value={value}
                onChange={(e) => setValue(e.target.value)}
                placeholder="Enter value"
                style={inputStyle}
              />
            </div>
          )}
          <button onClick={handleExecute} disabled={!connected}>
            Execute
          </button>
        </div>
        <div>
          <div style={{ fontWeight: "bold", marginBottom: "10px" }}>Reply</div>
          <div
            style={{
              backgroundColor: "#444444",
              width: "14rem",
              height: "14rem",
              padding: "10px",
            }}
          >
            {reply}
          </div>
        </div>
      </div>
    </div>
  );
}
