import { useCallback, useState } from "react";

const useLocalStorage = (key): [string | null, (string) => void] => {
  const [value, setValue] = useState(localStorage.getItem(key));
  const setItem = useCallback(
    (newValue) => {
      try {
        if (newValue) {
          localStorage.setItem(key, newValue);
        } else {
          localStorage.removeItem(key);
        }
        setValue(newValue);
      } catch (err) {
        console.error(
          `Error setting setting value for local storage key [${key}]`,
          err
        );
      }
    },
    [key]
  );
  return [value, setItem];
};

export default useLocalStorage;
