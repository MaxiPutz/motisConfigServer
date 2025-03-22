import {useMemo, useState } from "react";

interface SearchableDropdownProps {
    label: string;
    options: string[];
    value: string;
    onChange: (val: string) => void;
}
export const SearchableDropdown: React.FC<SearchableDropdownProps> = ({
    label,
    options,
    value,
    onChange,
}) => {
    const [search, setSearch] = useState("");

    const filteredOptions = useMemo(
        () =>
            options.filter((opt) =>
                opt.toLowerCase().includes(search.toLowerCase())
            ),
        [options, search]
    );


    return (
        <div style={{ margin: "0.5rem 0" }}>
            <label>
                {label}:{" "}
                <input
                    type="text"
                    placeholder="Search..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    style={{ marginRight: "0.5rem" }}
                />
            </label>
            <select value={value} onChange={(e) => onChange(e.target.value)}>
                <option value="">All</option>
                {filteredOptions.map((opt) => (
                    <option key={opt} value={opt}>
                        {opt}
                    </option>
                ))}
            </select>
        </div>
    );
};

