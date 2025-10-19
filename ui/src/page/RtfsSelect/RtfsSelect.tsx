// RtfsSelect.tsx
import { useState, ChangeEvent } from 'react';
//import data from "../../assets/gtfs.json";
import "./RtfsSelect.css";
import { useNavigate } from 'react-router';
import { NavArea } from '../component/NavArea';
import { useConfig } from '../../provider/configProvider';

export interface Transitous {
    Name: string;
    Url: string;
}

const groupByCountry = (files: Transitous[]) => {
    return files.reduce((groups: { [key: string]: Transitous[] }, file) => {
        const parts = file.Name.split('_');
        const country = parts[0];
        if (!groups[country]) {
            groups[country] = [];
        }
        groups[country].push(file);
        return groups;
    }, {});
};

export function RtfsSelect({ feeds  }: { feeds: Transitous[]; }) {
    const { store, setFeeds } = useConfig();

    console.log(feeds)

    const [selectedFiles, setSelectedFiles] = useState<Transitous[]>(store.feeds);
    const [expandedCountries, setExpandedCountries] = useState<Set<string>>(new Set());

    const nav = useNavigate();

    const groups = groupByCountry(feeds as Transitous[]);

    // Toggle a nation's expansion (show/hide its file list)
    const toggleCountry = (country: string) => {
        const newSet = new Set(expandedCountries);
        if (newSet.has(country)) {
            newSet.delete(country);
        } else {
            newSet.add(country);
        }
        setExpandedCountries(newSet);
    };

    // Toggle individual file selection. If already selected, remove it; if not, add it.
    const toggleFileSelection = (file: Transitous) => {
        if (selectedFiles.some(f => f.Name === file.Name)) {
            setSelectedFiles(selectedFiles.filter(f => f.Name !== file.Name));
            setFeeds(selectedFiles.filter(f => f.Name !== file.Name));
        } else {
            setSelectedFiles([...selectedFiles, file]);
            setFeeds([...selectedFiles, file]);
        }
    };

    // Handler for "Select All" checkbox for a country.
    const handleSelectAll = (country: string, checked: boolean) => {
        if (checked) {
            // Add all files for that country which are not already selected.
            const newFiles = groups[country].filter(
                file => !selectedFiles.some(f => f.Name === file.Name)
            );
            setSelectedFiles([...selectedFiles, ...newFiles]);
            setFeeds([...selectedFiles, ...newFiles]);
        } else {
            // Remove all files for that country.
            setSelectedFiles(
                selectedFiles.filter(file => file.Name.split('_')[0] !== country)
            );
            setFeeds(selectedFiles.filter(file => file.Name.split('_')[0] !== country));
        }
    };

    // Compute if all files in a group are selected.
    const isAllSelected = (country: string): boolean => {
        return groups[country].every(file => selectedFiles.some(f => f.Name === file.Name)
        );
    };

    // Format file name for display
    const formatFileName = (file: Transitous) => {
        const parts = file.Name.split('_');
        if (parts.length > 1) {
            const nameWithoutCountry = parts.slice(1).join('_');
            return nameWithoutCountry.replace(/\.gtfs\.zip$/, '');
        }
        return file.Name;
    };

    return (
        <div style={{ padding: 20, fontFamily: 'sans-serif' }}>
            <h2>Selected Files</h2>
            <div style={{ marginBottom: 20, border: '1px solid #ccc', padding: 10 }}>
                {selectedFiles.length === 0 ? (
                    <em>No files selected yet.</em>
                ) : (
                    selectedFiles.map(file => (
                        <div key={file.Name} style={{ padding: '5px 0' }}>
                            <strong>{formatFileName(file)}</strong> (
                            <a href={file.Url} target="_blank" rel="noopener noreferrer">
                                Download
                            </a>
                            )
                        </div>
                    ))
                )}
            </div>

            <h2>Countries</h2>
            {Object.keys(groups)
                .sort()
                .map(country => (
                    <div key={country} style={{ marginBottom: 10, border: '1px solid #ddd', padding: 10 }}>
                        <div style={{ cursor: 'pointer' }} onClick={() => toggleCountry(country)}>
                            <strong>{country}</strong> {expandedCountries.has(country) ? '▲' : '▼'}
                        </div>
                        <div style={{ padding: '2px 0' }}>
                            <label>
                                Select All
                                <input
                                    type="checkbox"
                                    checked={isAllSelected(country)}
                                    onChange={(e: ChangeEvent<HTMLInputElement>) => handleSelectAll(country, e.target.checked)}
                                    style={{ marginLeft: 5 }} />
                            </label>
                        </div>
                        {expandedCountries.has(country) && (
                            <div style={{ marginLeft: 20, marginTop: 5 }}>
                                {groups[country].map(file => (
                                    <div
                                        key={file.Name}
                                        style={{ display: 'flex', justifyContent: 'space-between', padding: '2px 0' }}
                                    >
                                        <span>{formatFileName(file)}</span>
                                        <label>
                                            <input
                                                type="checkbox"
                                                checked={selectedFiles.some(f => f.Name === file.Name)}
                                                onChange={() => toggleFileSelection(file)} />
                                        </label>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                ))}
            <NavArea handleNext={() => nav("/overview")} handlePrev={() => nav("/osm")} />
        </div>
    );
}
