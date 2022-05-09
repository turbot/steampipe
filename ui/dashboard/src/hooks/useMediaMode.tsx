import { useEffect, useState } from "react";

export type MediaMode = "screen" | "print";

const useMediaMode = () => {
  const [mediaMode, setMediaMode] = useState<MediaMode>("screen");
  useEffect(() => {
    const mediaQuery = window.matchMedia("print");
    const changeHandler = (e) => {
      if (e.matches) {
        setMediaMode("print");
      } else {
        setMediaMode("screen");
      }
    };
    mediaQuery.addEventListener("change", changeHandler);
    return () => mediaQuery.removeEventListener("change", changeHandler);
  }, []);
  return mediaMode;
};

export default useMediaMode;
