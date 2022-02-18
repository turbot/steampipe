import { useEffect, useState } from "react";

const useMediaMode = () => {
  const [mediaMode, setMediaMode] = useState("screen");
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
