// src/components/ProgressBar.tsx

import { useEffect } from "react";
import { useProgress } from "../../provider/progressProvider";


export const ProgressBar = ({setDownloadStated} : {setDownloadStated: (state :boolean) => void}) => {
  const { progressList } = useProgress(); // Get progressList from context

  progressList.sort((e1, e2) => e1.name.localeCompare(e2.name))


  useEffect(()=>{
    if (progressList.length != 0) {
        setDownloadStated(true)
    }
  }, [progressList])
  
  return (
    <div style={{background: "gray"}}>
      {progressList.length > 0 ? (
        progressList.map((progress, index) => (
          <div key={index} style={{ marginBottom: '10px' }}>
            <p>{progress.name}: {progress.data}%</p>
            <progress value={progress.data} max="100" style={{ width: '100%' }} />
          </div>
        ))
      ) : (
        <p>No active downloads.</p>
      )}
    </div>
  );
};