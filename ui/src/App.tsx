import './App.css'
import motisReleaseAsset from "../../assets/motis.json"

import { MotisSelect, } from './page/Motis/MotisReleaseSelect'
import { useEffect, useState } from 'react'
import { SelectOsm } from './page/selectOsm/SelectOsm'
import { BrowserRouter, Route, Routes } from 'react-router'
import { RtfsSelect } from './page/RtfsSelect/RtfsSelect'
import { ConfigOverview } from './page/Overview/Overview'
import { ConfigProvider } from './provider/configProvider'
import { useProgress } from './provider/progressProvider'
import { BusIcon } from 'lucide-react'


const protocol = window.location.protocol; // "http:" or "https:"
const wsProtocol = protocol === "https:" ? "wss:" : "ws:";
const host = window.location.host; // "example.com:3000" or "localhost:5173"

export const ENV = {
  baseURL:
    import.meta.env.VITE_API_URL === "/"
      ? `${protocol}//${host}`
      : import.meta.env.VITE_API_URL,

  baseWsUrl:
    import.meta.env.VITE_ES_URL === "/"
      ? `${wsProtocol}//${host}`
      : import.meta.env.VITE_ES_URL,
};

export interface ProgressBar {
  name: string,
  data: number
}

function updateProgressBar(store: ProgressBar[], update: ProgressBar) {
  return store.reduce((prev, cur) => cur.name == update.name ? [...prev] : [...prev, cur], [update] as ProgressBar[])
}

// New Header Component
function Header() {
  return (
    <header className="header">
      <div className="logo">
        <BusIcon />
        <h1>Transit Configurator</h1>
      </div>
    </header>
  )
}

function App() {
  const [motisRelease, setMotisRelease] = useState(motisReleaseAsset)
  const [osAndArch, setOsAndArch] = useState({ os: "", arch: "" })
  const { setProgressBar } = useProgress()

  let progresses: ProgressBar[] = []




  useEffect(() => {

    

    const socket = new WebSocket(`${ENV.baseWsUrl}/ws/`);

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      console.log("Received:", data);

      if (data.name && data.data) {
        if (data.name === "terminaldata") {
          if (data.data.includes("%")) {
            const progressNumber = Number(data.data.split("]")[data.data.split("]").length - 1].split("%")[0].trim())
            const name = data.data.split(":")[0].split("[K")[1].replaceAll(" ", "")
            progresses = updateProgressBar(progresses, { name: name, data: progressNumber })
            progresses = updateProgressBar(progresses, data)
            setProgressBar(progresses)
          }
        } else {
          progresses = updateProgressBar(progresses, data)
          setProgressBar(progresses)
        }
      }
    };


    fetch(`${ENV.baseURL}/init`).then((res) => res.json())
      .then(json => {
        setMotisRelease(json.releases)
        setOsAndArch({
          arch: json.arch,
          os: json.os
        })
      })
  }, [])

  return (
    <>
      <Header />
      <ConfigProvider>
        <div className="container">
          <BrowserRouter>
            <Routes>
              <Route path='/' element={<MotisSelect updatedReleases={motisRelease} osAndArch={osAndArch} />} />
              <Route path='/osm' element={<SelectOsm />} />
              <Route path='/rtfs' element={<RtfsSelect />} />
              <Route path='/overview' element={<ConfigOverview />} />
            </Routes>
          </BrowserRouter>
        </div>
      </ConfigProvider>
    </>
  )
}

export default App
