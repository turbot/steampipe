import { useEffect, useState } from "react";

const useDashboardIcons = () => {
  const [heroIcons, setHeroIcons] = useState<any>({});
  const [materialSymbols, setMaterialSymbols] = useState<any>({});

  // Dynamically import hero icons from its own bundle
  useEffect(() => {
    import("../icons/heroIcons").then((m) => setHeroIcons(m.icons));
  }, []);

  // Dynamically import material symbols from its own bundle
  useEffect(() => {
    import("../icons/materialSymbols").then((m) => setMaterialSymbols(m.icons));
  }, []);

  return {
    heroIcons,
    materialSymbols,
  };
};

export default useDashboardIcons;
