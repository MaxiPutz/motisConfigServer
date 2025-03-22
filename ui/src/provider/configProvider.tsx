import  {
  createContext,
  useContext,
  useState,
  ReactNode,
  FC,
} from 'react';
import { Release } from '../page/Motis/MotisReleaseSelect';
import { Transitous } from '../page/RtfsSelect/RtfsSelect';

export interface Store {
  feeds: Transitous[];
  osmUrl: string;
  motisUrl: Release;
}

interface ConfigContextType {
  store: Store;
  setFeeds: (feeds: Transitous[]) => void;
  setOsmUrl: (osmUrl: string) => void;
  setMotisUrl: (motisUrl: Release) => void;
}

// Default config (adjust as needed)
const defaultStore: Store = {
  feeds: [],
  osmUrl: '',
  motisUrl: { arch: "", browser_download_url: "", name: "", os: "", tag_name: "" },
};

const ConfigContext = createContext<ConfigContextType | undefined>(undefined);

interface ConfigProviderProps {
  children: ReactNode;
}

export const ConfigProvider: FC<ConfigProviderProps> = ({ children }) => {
  const [store, setStore] = useState<Store>(defaultStore);

  const setFeeds = (feeds: Transitous[]) => {
    setStore((prev) => ({ ...prev, feeds }));
  };

  const setOsmUrl = (osmUrl: string) => {
    setStore((prev) => ({ ...prev, osmUrl }));
  };

  const setMotisUrl = (motisUrl: Release) => {
    setStore((prev) => ({ ...prev, motisUrl }));
  };

  return (
    <ConfigContext.Provider value={{ store, setFeeds, setOsmUrl, setMotisUrl }}>
      {children}
    </ConfigContext.Provider>
  );
};

// Custom hook to access the config context
export const useConfig = (): ConfigContextType => {
  const context = useContext(ConfigContext);
  if (!context) {
    throw new Error('useConfig must be used within a ConfigProvider');
  }
  return context;
};
