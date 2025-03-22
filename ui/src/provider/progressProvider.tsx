import { createContext, useState, ReactNode,  useContext } from 'react';

interface ProgressBar {
    name: string;
    data: number;
}

// Define the context outside the component
export const ProgressContext = createContext<{
    progressList: ProgressBar[];
    setProgressBar: (progressBarList: ProgressBar[]) => void;
}>({
    progressList: [],
    setProgressBar: () => { },
});

export function ProgressProvider({ children }: { children: ReactNode }) {
    const [progressList, setProgressBar] = useState<ProgressBar[]>([]);

    return (
        <ProgressContext.Provider value={{ progressList, setProgressBar }}>
            {children} {/* Render children to pass context down */}
        </ProgressContext.Provider>
    );
}


export function useProgress() {
    const progress = useContext(ProgressContext)


    return progress
}