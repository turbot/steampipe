import React, { useLayoutEffect, useRef, useState } from "react";

const useDimensions = (): readonly [
  React.RefObject<HTMLDivElement>,
  DOMRect
] => {
  const ref = useRef<HTMLDivElement>(null);
  // Initialize state with undefined width/height so server and client renders match
  // Learn more here: https://joshwcomeau.com/react/the-perils-of-rehydration/
  const [dimensions, setDimensions] = useState({} as DOMRect);

  useLayoutEffect(() => {
    // Handler to call on window resize
    function handleResize() {
      // Set window width/height to state
      // @ts-ignore
      setDimensions(ref.current.getBoundingClientRect().toJSON());
    }

    if (!ref.current) {
      return;
    }

    // Remove event listener
    window.removeEventListener("resize", handleResize);
    // Add event listener
    window.addEventListener("resize", handleResize);
    // Call handler right away so state gets updated with initial window size
    handleResize();
    // Remove event listener on cleanup
    return () => window.removeEventListener("resize", handleResize);
  }, [ref.current]); // Empty array ensures that effect is only run on mount

  return [ref, dimensions];
};

export default useDimensions;
