import { useNavigate } from "react-router-dom";

export const urlQueryParamHistoryMode = {
  PUSH: "push",
  REPLACE: "replace",
};

type FallbackValue = string | null;

// eslint-disable-next-line import/no-anonymous-default-export
const useQueryParams = (
  key: string,
  fallbackValue: FallbackValue = null,
  historyMode: string = urlQueryParamHistoryMode.PUSH
): readonly [string | null, (string) => void] => {
  const navigate = useNavigate();
  const searchParams = new URLSearchParams(window.location.search);
  const queryValue = searchParams.get(key) || fallbackValue;

  const setQueryValue = (newValue) => {
    const setSearchParams = new URLSearchParams(window.location.search);
    if (newValue) {
      setSearchParams.set(key, newValue);
    } else {
      setSearchParams.delete(key);
    }
    if (historyMode === urlQueryParamHistoryMode.REPLACE) {
      navigate(`${window.location.pathname}?${setSearchParams.toString()}`, {
        replace: true,
      });
    } else {
      navigate(`${window.location.pathname}?${setSearchParams.toString()}`);
    }
  };
  return [queryValue, setQueryValue];
};

export default useQueryParams;
