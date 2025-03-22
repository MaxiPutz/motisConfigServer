import { useEffect, useState } from "react"
import { SearchableDropdown } from "./SearchDorpDown/SearchDropDown"
import { useNavigate } from "react-router"
import { NavArea } from "../component/NavArea"
import { useConfig } from "../../provider/configProvider"


export interface Release {
    name: string,
    os: string,
    arch: string,
    browser_download_url: string,
    tag_name: string,
}

interface Assets {
    name: string,
    browser_download_url: string
    os: string,
    arch: string
}

export interface ReleaseGithubStruct {
    name: string,
    tag_name: string,
    assets: Assets[]
}

export function MotisSelect({ updatedReleases, osAndArch }: { updatedReleases: ReleaseGithubStruct[], osAndArch: { os: string, arch: string } }) {
    const { setMotisUrl, store } = useConfig()

    const [os, setOs] = useState(osAndArch.os)
    const [arch, setArch] = useState(osAndArch.arch)
    const [version, setVersion] = useState("")
    const [selectedDownload, setSelectedDownload] = useState<Release>(store.motisUrl)
    const [releases, setReleases] = useState<ReleaseGithubStruct[]>(updatedReleases)
    
    const nav = useNavigate()



    useEffect(() => {
        console.log(osAndArch, "osAndArch");

        if (osAndArch.os != "") {
            console.log(osAndArch, "osAndArch");

            setOs(osAndArch.os)
        }
        if (osAndArch.arch != "") {
            if (osAndArch.os === "windows") {
                setArch("")
            } else {
                setArch(osAndArch.arch)
            }
        }
    }, [osAndArch])

    useEffect(() => {
        setReleases(updatedReleases)
    }, [updatedReleases])

    const osOptions = releases.reduce((prev, cur) => {
        cur.assets.forEach(e => {
            if (!prev.some(p => e.os == p)) {
                prev.push(e.os)
            }
        })
        return [...prev]
    }, [] as string[])

    const archOptions = releases.reduce((prev, cur) => {
        cur.assets.forEach(e => {
            if (!prev.some(p => e.arch == p)) {
                prev.push(e.arch)
            }
        })
        return [...prev]
    }, [] as string[])

    const data = releases.reduce((prev, cur) =>
        [...prev, ...cur.assets.map(e => ({
            ...e,
            tag_name: cur.tag_name
        }))]
        , [] as Release[])


    const vOptions = releases.reduce((prev, cur) => prev.some(p => cur.tag_name == p) ? prev : [...prev, cur.tag_name], [] as string[])


    console.log(osOptions);
    console.log(archOptions);
    console.log(vOptions);
    console.log(data);



    return (
        <>

            <h1>
                Selected
                {os}
                {arch}
            </h1>
            <h3>
                {selectedDownload?.browser_download_url}
            </h3>
            <SearchableDropdown
                label="os"
                onChange={(str) => setOs(str)}
                options={osOptions}
                value={os}
            />
            <SearchableDropdown
                label="arch"
                onChange={(str) => setArch(str)}
                options={archOptions}
                value={arch}
            />
            <SearchableDropdown
                label="version"
                onChange={(str) => setVersion(str)}
                options={vOptions}
                value={version}
            />

            {
                data
                    .filter(e => arch == "" ? true : e.arch == arch)
                    .filter(e => os == "" ? true : e.os == os)
                    .filter(e => version == "" ? true : e.tag_name == version)
                    .map((e, i) => <div key={i}>
                        <label>
                            <input type="radio" checked={e.browser_download_url === selectedDownload?.browser_download_url} onChange={() => {
                                setSelectedDownload(e)
                                setMotisUrl(e)
                            }} ></input>
                            {
                                e.browser_download_url
                            }
                        </label>
                    </div>)
            }
            <NavArea handleNext={() => nav("/osm")} handlePrev={() => undefined} />
        </>
    )
}

