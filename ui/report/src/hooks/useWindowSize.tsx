import React, { useEffect, useState } from "react";

// https://stackoverflow.com/questions/19014250/rerender-view-on-browser-resize-with-react
const useWindowSize = (): readonly [number, number] => {
  // @ts-ignore
  const [size, setSize] = useState(
    typeof window !== "undefined"
      ? ([window.innerWidth, window.innerHeight] as const)
      : ([0, 0] as const)
  );
  useEffect(() => {
    const updateSize = () => {
      setSize([window.innerWidth, window.innerHeight]);
    };
    window.addEventListener("resize", updateSize);
    return () => window.removeEventListener("resize", updateSize);
  }, []);
  return size;
};

export default useWindowSize;
