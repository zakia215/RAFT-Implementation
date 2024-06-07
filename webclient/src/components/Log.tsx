import { useState, useEffect } from "react";
import { log } from "../lib/lib";

type LogProps = {
  connected: boolean;
};

export function Log({ connected }: LogProps) {
  const [currentLog, setCurrentLog] = useState<React.ReactNode[]>([]);

  useEffect(() => {
    let interval: number;

    if (connected) {
      interval = setInterval(async () => {
        try {
          const newLog = await log();
          setCurrentLog(
            newLog.map((log, index) => (
              <div
                key={index}
              >{`Term: ${log.term}, Command: ${log.command.command} ${log.command.key} ${log.command.value}`}</div>
            ))
          );
        } catch (error) {
          console.error("Failed to fetch log:", error);
        }
      }, 1000);
    }

    return () => {
      if (interval) {
        clearInterval(interval);
      }
    };
  }, [connected]);
  return (
    <div>
      <div style={{ fontWeight: "bold", marginBottom: "10px" }}>Log</div>
      <div
        style={{
          backgroundColor: "black",
          width: "50vw",
          height: "80vh",
          padding: "10px",
          overflow: "auto",
        }}
      >
        {currentLog}
      </div>
    </div>
  );
}
