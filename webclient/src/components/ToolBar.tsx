import { useState } from "react";
import { connect, ping, get, set, strln, append, del } from "../lib/lib";
import { toast } from "react-toastify";

type ToolBarProps = {
  connected: boolean;
  setConnected: React.Dispatch<React.SetStateAction<boolean>>;
};

export function ToolBar({ connected, setConnected }: ToolBarProps) {
  const [ip, setIp] = useState("");
  const [port, setPort] = useState("");
  const [selectedCommand, setSelectedCommand] = useState("Select command");
  const [key, setKey] = useState("");
  const [value, setValue] = useState("");
  const [reply, setReply] = useState("");

  const commandOptions = ["ping", "get", "set", "strln", "append", "del"];

  const handleConnect = async () => {
    try {
      toast.loading("Connecting to server");
      const success = await connect(ip, port);
      setConnected(success);
      toast.dismiss();
      toast.success(`Connected to ${ip}:${port}`, { autoClose: 1000 });
    } catch (error) {
      toast.dismiss();
      toast.error("Failed to connect to server");
    }
  };

  const handleExecute = async () => {
    let commandReply = "";
    toast.loading("Executing command...");
    try {
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
        case "del":
          commandReply = await del(key);
          break;
        default:
          toast.dismiss();
          toast.error("Invalid command");
      }
      toast.dismiss();
      toast.success("Command executed successfully", { autoClose: 1000 }  );
    } catch (error) {
      toast.dismiss();
      toast.error("Failed to execute command");
    }
    setReply(commandReply);
  };

  const labelStyle = {
    display: "block",
    fontSize: "0.9em",
    marginBottom: "8px",
  };

  const inputStyle: React.CSSProperties = {
    paddingLeft: "8px",
    paddingRight: "8px",
    paddingTop: "6px",
    paddingBottom: "6px",
    fontSize: "0.9em",
    fontFamily: "inherit",
    borderRadius: "6px",
    backgroundColor: "#444444",
    border: "none",
  };

  return (
    <div style={{ paddingLeft: "40px" }}>
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          marginBottom: "12px",
        }}
      >
        <label style={labelStyle}>IP Address</label>
        <input
          type="text"
          value={ip}
          onChange={(e) => setIp(e.target.value)}
          placeholder="Enter IP address"
          style={inputStyle}
        />
      </div>
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          marginBottom: "16px",
        }}
      >
        <label style={labelStyle}>Port</label>
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
          gap: "20px",
          marginTop: "20px",
        }}
      >
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            gap: "12px",
          }}
        >
          <div
            style={{
              display: "flex",
              flexDirection: "column",
            }}
          >
            <label style={labelStyle}>Command</label>
            <select
              value={selectedCommand}
              onChange={(e) => setSelectedCommand(e.target.value)}
              style={{
                padding: "4px",
                fontSize: "0.9em",
                fontFamily: "inherit",
                borderRadius: "6px",
                backgroundColor: "#444444",
              }}
            >
              {commandOptions.map((choice, index) => (
                <option key={index} value={choice}>
                  {choice}
                </option>
              ))}
            </select>
          </div>
          {selectedCommand && selectedCommand !== "ping" && (
            <div
              style={{
                display: "flex",
                gap: "24px",
                alignItems: "center",
              }}
            >
              <label style={{
                display: "block",
                fontSize: "0.9em",
                marginBottom: "2px",
              }}>Key</label>
              <input
                type="text"
                value={key}
                onChange={(e) => setKey(e.target.value)}
                placeholder="Enter key"
                style={{
                  paddingLeft: "8px",
                  paddingRight: "8px",
                  paddingTop: "6px",
                  paddingBottom: "6px",
                  fontSize: "0.9em",
                  fontFamily: "inherit",
                  borderRadius: "6px",
                  backgroundColor: "#444444",
                  border: "none",
                  width: "7rem",
                }}
              />
            </div>
          )}
          {(selectedCommand === "set" || selectedCommand === "append") && (
            <div
              style={{
                display: "flex",
                gap: "12px",
                alignItems: "center",
              }}
            >
              <label style={{
                display: "block",
                fontSize: "0.9em",
                marginBottom: "2px",
              }}>Value</label>
              <input
                type="text"
                value={value}
                onChange={(e) => setValue(e.target.value)}
                placeholder="Enter value"
                style={{
                  paddingLeft: "8px",
                  paddingRight: "8px",
                  paddingTop: "6px",
                  paddingBottom: "6px",
                  fontSize: "0.9em",
                  fontFamily: "inherit",
                  borderRadius: "6px",
                  backgroundColor: "#444444",
                  border: "none",
                  width: "7rem",
                }}
              />
            </div>
          )}
          <button 
            onClick={handleExecute} 
            disabled={!connected}
            title={!connected ? "Connect to server first" : ""}
            style={{
              marginTop: "8px",
            }}
          >
            Execute
          </button>
        </div>
        {connected ? (
          <div>
            <div style={{ 
              fontSize: "0.9em", 
              marginBottom: "8px" }}>Server Response</div>
            <div
              style={{
                backgroundColor: "#2e2e2e",
                width: "16rem",
                height: "16rem",
                padding: "8px",
                borderRadius: "6px",
                fontFamily: "monospace",
              }}
            >
              <span className="flicker">$</span> {reply}
            </div>
          </div>
        ) : (
          <div 
            style={{
              width: "16rem",
              height: "16rem",
              padding: "8px",
              borderRadius: "6px",
            }}
          />
        )}
      </div>
    </div>
  );
}
