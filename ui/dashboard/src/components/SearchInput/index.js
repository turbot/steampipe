import { classNames } from "../../utils/styles";
import { ClearIcon, SearchIcon } from "../../constants/icons";
import { forwardRef } from "react";

const SearchInput = forwardRef(
  (
    {
      className,
      disabled = false,
      placeholder,
      readOnly = false,
      setValue,
      value,
    },
    ref
  ) => {
    return (
      <div className="relative">
        <div className="pointer-events-none absolute inset-y-0 left-0 pl-3 flex items-center text-foreground-light text-sm">
          <SearchIcon className="h-4 w-4" />
        </div>
        <input
          className={classNames(
            className,
            "flex-1 block w-full bg-dashboard-panel rounded-md border border-table-divide px-8 overflow-x-auto text-sm md:text-base disabled:bg-black-scale-1"
          )}
          disabled={disabled}
          onChange={(e) => setValue(e.target.value)}
          placeholder={placeholder}
          readOnly={readOnly}
          ref={ref}
          type="text"
          value={value}
        />
        {value && (
          <div
            className="absolute inset-y-0 right-0 pr-3 flex items-center cursor-pointer text-foreground"
            onClick={() => setValue("")}
          >
            <ClearIcon className="h-4 w-4" />
          </div>
        )}
      </div>
    );
  }
);

export default SearchInput;
