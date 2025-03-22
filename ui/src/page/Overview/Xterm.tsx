import { Terminal } from "@xterm/xterm";
import { useEffect, useRef } from "react";
import { useProgress } from "../../provider/progressProvider";
import "@xterm/xterm/css/xterm.css"; // Import Xterm styles

export function Xterm() {
  const xtermRefContainer = useRef<HTMLDivElement | null>(null);
  const xtermRef = useRef<Terminal | null>(null);
  const { progressList } = useProgress();

  // Initialize the terminal
  useEffect(() => {
    if (!xtermRefContainer.current) return;

    // Create a new terminal instance if it doesn't exist
    if (!xtermRef.current) {
      xtermRef.current = new Terminal({
        convertEol: true, // Convert newline characters to CRLF
        disableStdin: true, // Disable user input
        cursorBlink: false, // Disable cursor blinking
        allowProposedApi: true, // Enable proposed APIs (if needed)
      });
      xtermRef.current.open(xtermRefContainer.current);
    }

    // Cleanup on unmount
    return () => {
      if (xtermRef.current) {
        xtermRef.current.dispose();
        xtermRef.current = null;
      }
    };
  }, []);

  // Update the terminal with new data from progressList
  useEffect(() => {
    if (!xtermRef.current) return;

    // Find the terminal data in progressList
    const data = progressList.find((e) => e.name === "terminaldata");
    if (data && typeof data.data === "string") {
      console.log("Raw data:", data.data); // Debug: Log the raw data
      xtermRef.current.write(data.data)
    }
  }, [progressList]);

  return (
    <div
      ref={xtermRefContainer}
      style={{ width: "100%", height: "100%", padding: "10px" }}
    />
  );
}