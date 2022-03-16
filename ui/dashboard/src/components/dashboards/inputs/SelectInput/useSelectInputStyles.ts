import { useDashboard } from "../../../../hooks/useDashboard";
import { useEffect, useState } from "react";

const useSelectInputStyles = () => {
  const [, setRandomVal] = useState(0);
  const {
    themeContext: { theme, wrapperRef },
  } = useDashboard();

  // This is annoying, but unless I force a refresh the theme doesn't stay in sync when you switch
  useEffect(() => setRandomVal(Math.random()), [theme.name]);

  if (!wrapperRef) {
    return null;
  }

  // @ts-ignore
  const style = window.getComputedStyle(wrapperRef);
  const background = style.getPropertyValue("--color-background");
  const backgroundPanel = style.getPropertyValue("--color-background-panel");
  const foreground = style.getPropertyValue("--color-foreground");
  const blackScale3 = style.getPropertyValue("--color-black-scale-3");

  return {
    clearIndicator: (provided) => ({
      ...provided,
      cursor: "pointer",
    }),
    control: (provided, state) => {
      return {
        ...provided,
        backgroundColor: backgroundPanel,
        borderColor: state.isFocused ? "#2684FF" : blackScale3,
        boxShadow: "none",
      };
    },
    dropdownIndicator: (provided) => ({
      ...provided,
      cursor: "pointer",
    }),
    input: (provided) => {
      return {
        ...provided,
        color: foreground,
      };
    },
    singleValue: (provided) => {
      return {
        ...provided,
        color: foreground,
      };
    },
    menu: (provided) => {
      return {
        ...provided,
        backgroundColor: backgroundPanel,
        border: `1px solid ${blackScale3}`,
        boxShadow: "none",
        marginTop: 0,
        marginBottom: 0,
      };
    },
    menuList: (provided) => {
      return {
        ...provided,
        paddingTop: 0,
        paddingBottom: 0,
      };
    },
    option: (provided, state) => {
      return {
        ...provided,
        backgroundColor: state.isFocused ? background : "none",
        color: foreground,
      };
    },
  };
};

export default useSelectInputStyles;
