import { useCallback, useEffect } from "react";

const useDebouncedEffect = (effect, delay, deps) => {
  const callback = useCallback(effect, deps); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    const handler = setTimeout(() => {
      callback();
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [callback, delay]);
};

export default useDebouncedEffect;
