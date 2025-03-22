import React, { useEffect, useState } from 'react';
import geofabirk from "../../assets/geofabrik.json"
import { useNavigate } from 'react-router';
import { NavArea } from '../component/NavArea';
import { useConfig } from '../../provider/configProvider';

// Define the tree node interface.
export interface GeofabrikTreeNode {
    path: string;
    depth: number;
    children: GeofabrikTreeNode[];
    isLeaf: boolean;
    regionName: string;
    osmData: string;
    rawHtml: string;
    rootPath: string;
}

function getGeofabrikAsList(node: GeofabrikTreeNode, result: GeofabrikTreeNode[]): GeofabrikTreeNode[] {
    result.push(node)
    if (node.isLeaf || node.children === null) {
        return result
    }
    for (let ele of node.children) {
        console.log(ele);

        getGeofabrikAsList(ele, result)
    }

    return result
}


export const SelectOsm: React.FC = () => {

    const { store, setOsmUrl } = useConfig()

    const [tree, setTree] = useState<GeofabrikTreeNode | null>(null);
    const [selectedOsm, setSelectedOsm] = useState<string>(store.osmUrl);
    const [selectedOsmNode, setSelectedOsmNode] = useState<GeofabrikTreeNode>();

    const nav = useNavigate()

    const handleSelect = (selectedOsm: string, selectedOsmNode: GeofabrikTreeNode) => {
        setSelectedOsm(selectedOsm)
        setSelectedOsmNode(selectedOsmNode)
        setOsmUrl(selectedOsm)
    }

    const computeDownloadUrl = (node: GeofabrikTreeNode): string => {
        if (node.rawHtml && node.rootPath) {
            // Remove the ".html" extension from rootPath and append osmData.
            const base = node.rootPath.replace(/\.html$/, '');
            return `${node.rawHtml}${base}/${node.osmData}`;
        } else {
            return `${node.rawHtml}/${node.osmData}`
        }
        return node.osmData;
    };

    useEffect(() => {
        console.log(geofabirk);

        setTree(geofabirk as unknown as GeofabrikTreeNode);
        let result = [] as GeofabrikTreeNode[]
        console.log("list", getGeofabrikAsList(geofabirk as unknown as GeofabrikTreeNode, result))
        console.log("list result ", result)

    }, []);


    return (
        <div style={{ padding: '1rem' }}>
            <h1>Geofabrik File Explorer</h1>
            <h2>Selected OSM File:</h2>
            {
                selectedOsmNode ?
                    <div>

                        {selectedOsm ? <p>{selectedOsm}</p> : <p>None selected</p>}
                        {selectedOsmNode ? <p>{computeDownloadUrl(selectedOsmNode)}</p> : <p>None selected</p>}
                    </div> : <></>
            }
            <div style={{ marginBottom: '1rem' }}>

            </div>
            {(tree ?
                tree.children.map((e, i) => <div key={i}>
                    <TreeView
                        isDefaultOpen={false}
                        node={e}
                        selectedOsm={selectedOsm}
                        onSelect={handleSelect}
                    />
                </div>)
                : <></>
            )
            }
            <div style={{ marginTop: '1rem' }}>
            </div>
            {/* Navigation Footer */}
            <NavArea handleNext={() => nav("/rtfs")} handlePrev={() => nav("/")} />

        </div>
    );
};

interface TreeViewProps {
    node: GeofabrikTreeNode;
    selectedOsm: string;
    isDefaultOpen: boolean;
    onSelect: (osm: string, osmNode: GeofabrikTreeNode) => void;
}

const TreeView: React.FC<TreeViewProps> = ({ node, selectedOsm, isDefaultOpen, onSelect }) => {
    const [expanded, setExpanded] = useState<boolean>(isDefaultOpen);



    return (
        <div style={{ marginLeft: `${node.depth * 20}px`, marginBottom: '4px' }}>
            <div style={{ display: 'flex', alignItems: 'center' }}>
                {node.children.length > 0 && (
                    <button
                        onClick={() => setExpanded(!expanded)}
                        style={{ marginRight: '8px' }}
                    >
                        {expanded ? '-' : '+'}
                    </button>
                )}
                {true ? (
                    <label>
                        <input
                            type="radio"
                            name="osmFile"
                            value={node.osmData}
                            checked={selectedOsm === node.osmData}
                            onChange={() => onSelect(node.osmData, node)}
                            style={{ marginRight: '4px' }}
                        />
                        {(node.regionName)}
                    </label>
                ) : (
                    <span>{node.regionName}</span>
                )}
            </div>
            {expanded && node.children.map(child => (
                <TreeView
                    isDefaultOpen={false}
                    key={child.path}
                    node={child}
                    selectedOsm={selectedOsm}
                    onSelect={onSelect}
                />
            ))}
        </div>
    );
};

