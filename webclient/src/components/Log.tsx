import { useState, useEffect, useRef } from "react";
import { log } from "../lib/lib";

type LogProps = {
  connected: boolean;
};

export function Log({ connected }: LogProps) {
  const [currentLog, setCurrentLog] = useState<React.ReactNode[]>([]);
  const logDivRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (logDivRef.current) {
      logDivRef.current.scrollTop = logDivRef.current.scrollHeight;
    }
  }, [currentLog]);

  useEffect(() => {
    let interval: number;

    if (connected) {
      interval = setInterval(async () => {
        try {
          const newLog = await log();
          setCurrentLog(
            newLog.map((log, index) => (
              <div key={index} style={{ fontFamily: "monospace" }}>
                <span>{`$ `}</span>
                <span style={{ color: "lime" }}>{`Term: `}</span>
                <span>{`${log.term}`}</span>
                <span>{` | `}</span>
                <span style={{ color: "lime" }}>{`Command: `}</span>
                <span>{`${log.command.command} ${log.command.key} ${log.command.value}`}</span>
              </div>
            
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
      <div ref={logDivRef}
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
