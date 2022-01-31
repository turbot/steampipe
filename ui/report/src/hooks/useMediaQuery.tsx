import { useEffect, useState } from "react";

const useMediaQuery = (query) => {
  let mediaQuery;
  if (
    typeof window !== "undefined" &&
    typeof window.matchMedia !== "undefined"
  ) {
    mediaQuery = window.matchMedia(query);
  }

  const [match, setMatch] = useState(mediaQuery ? !!mediaQuery.matches : false);

  useEffect(() => {
    if (!mediaQuery) {
      return;
    }
    const handler = () => setMatch(!!mediaQuery.matches);
    mediaQuery.addEventListener("change", handler);
    return () => mediaQuery.removeEventListener("change", handler);
  }, []);

  return match;
};

export default useMediaQuery;
