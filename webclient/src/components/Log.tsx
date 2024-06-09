import { useState, useEffect, useRef } from "react";
import { log } from "../lib/lib";

type LogProps = {
  connected: boolean;
};

export function Log({ connected }: LogProps) {
  const [currentLog, setCurrentLog] = useState<React.ReactNode[]>([]);
  const logDivRef = useRef<HTMLDivElement | null>(null);
  const isUserAtBottomRef = useRef(true);

  useEffect(() => {
    const handleScroll = () => {
      if (!logDivRef.current) return;
      const { scrollTop, scrollHeight, clientHeight } = logDivRef.current;
      isUserAtBottomRef.current = scrollTop + clientHeight >= scrollHeight - 10;
    };

    const logDiv = logDivRef.current;
    logDiv?.addEventListener("scroll", handleScroll);
    
    return () => {
      logDiv?.removeEventListener("scroll", handleScroll);
    };
  }, []);

  useEffect(() => {
    if (isUserAtBottomRef.current && logDivRef.current) {
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
      <div style={{ marginBottom: "6px", fontSize: "0.9em" }}>Logs</div>
      <div
        ref={logDivRef}
        style={{
          backgroundColor: "#2e2e2e",
          width: "50vw",
          height: "80vh",
          padding: "12px",
          overflow: "auto",
          borderRadius: "6px",
        }}
      >
        {currentLog}
      </div>
    </div>
  );
}
