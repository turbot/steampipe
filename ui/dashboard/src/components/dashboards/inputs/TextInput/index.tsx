import { classNames } from "../../../../utils/styles";
import { DashboardActions, useDashboard } from "../../../../hooks/useDashboard";
import { debounce } from "lodash";
import { IInput, InputProps } from "../index";
import { useEffect, useMemo, useRef, useState } from "react";

const TextInput = (props: InputProps) => {
  const inputRef = useRef(null);
  const { dispatch, selectedDashboardInputs } = useDashboard();
  const stateValue = selectedDashboardInputs[props.name];
  const [value, setValue] = useState<string>(() => {
    return stateValue || "";
  });

  const changeHandler = (e) => {
    setValue(e.target.value);
  };

  const debouncedChangeHandler = useMemo(
    () => debounce(changeHandler, 400),
    []
  );

  // Cleanup
  useEffect(() => {
    return () => {
      debouncedChangeHandler.cancel();
    };
  }, []);

  useEffect(() => {
    dispatch({
      type: DashboardActions.SET_DASHBOARD_INPUT,
      name: props.name,
      value,
    });
  }, [dispatch, props.name, value]);

  useEffect(() => {
    if (!stateValue) {
      setValue("");
    }
  }, [stateValue]);

  return (
    <div>
      {props.properties.label && (
        <label htmlFor={props.name} className="block mb-1">
          {props.properties.label}
        </label>
      )}
      <div>
        <input
          ref={inputRef}
          type="text"
          name={props.name}
          id={props.name}
          className="block w-full sm:text-sm rounded-md border-black-scale-3"
          defaultValue={value}
          onChange={debouncedChangeHandler}
          placeholder={props.properties.placeholder}
        />
      </div>
    </div>
  );
};

const definition: IInput = {
  type: "text",
  component: TextInput,
};

export default definition;
