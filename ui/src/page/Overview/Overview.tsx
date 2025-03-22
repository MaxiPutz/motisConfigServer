import React, { useState } from 'react';
import { useNavigate } from 'react-router';
import { NavArea } from '../component/NavArea';
import { useConfig } from '../../provider/configProvider';
import { ENV } from '../../App';
import { ProgressBar } from './ProgressBar';

interface RequestDownload {
    gtfsUrls: string[]
    osmUrl: string
    motisUrl: string
}



export const ConfigOverview: React.FC = () => {
    const { store } = useConfig();
    const [isDownloadStated, setIsDownloadStarted] = useState(false)
    const nav = useNavigate()

    const handleNext = () => {

        const base = ENV.baseURL


        console.log(JSON.stringify({
            gtfsURLs: store.feeds.map(e => e.Url),
            motisUrl: store.motisUrl.browser_download_url,
            osmURLs: store.osmUrl
        } as unknown as RequestDownload));



        fetch(`${base}/startDownload`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                gtfsUrls: store.feeds.map(e => e.Url),
                motisUrl: store.motisUrl.browser_download_url,
                osmUrl: "https://download.geofabrik.de/" + store.osmUrl
            } as RequestDownload)
        })
    }


    return (
        <div style={{
            padding: '1rem',
            border: '1px solid #ccc',
            borderRadius: '5px',
            marginTop: '1rem',
            backgroundColor: '#f9f9f9',
            width: "90vw"
        }}>
            <h2>Selected Configuration</h2>
            <p>
                after the process is finished you can go to the out folder and run ./motis server
            </p>
            <p>
                <strong>Feeds:</strong> {store.feeds.length ? store.feeds.map(e => e.Name).join(', ') : 'None'}
            </p>
            <p>
                <strong>OS URL:</strong> {store.osmUrl || 'Not set'}
            </p>
            <p>
                <strong>Motis URL:</strong> {store.motisUrl.browser_download_url || 'Not set'}
            </p>

            <ProgressBar setDownloadStated={(state) => setIsDownloadStarted(state)} />
            {
                isDownloadStated ? <></> :
                    <NavArea handleNext={() => handleNext()} handlePrev={() => nav("/rtfs")} />
            }
        </div>
    );
};
