import React, { useState, useEffect, createContext, useContext } from "react";
import copy from "copy-to-clipboard";
import {
  CopyToClipboardIcon,
  CopyToClipboardSuccessIcon,
} from "../../constants/icons";
import { classNames } from "../../utils/styles";

interface ICopyToClipboardContext {
  doCopy: boolean;
  setDoCopy: (value: boolean) => void;
}

const CopyToClipboardContext = createContext<ICopyToClipboardContext | null>(
  null
);

const CopyToClipboardProvider = ({ children }) => {
  const [doCopy, setDoCopy] = useState(false);
  return (
    <CopyToClipboardContext.Provider value={{ doCopy, setDoCopy }}>
      {children({ setDoCopy })}
    </CopyToClipboardContext.Provider>
  );
};

const CopyToClipboard = ({
  data,
  className = "text-foreground-light",
  stopPropagation = true,
}) => {
  const context = useContext(CopyToClipboardContext);
  const { doCopy, setDoCopy } = context
    ? context
    : ({} as ICopyToClipboardContext);
  const [copySuccess, setCopySuccess] = useState(false);

  const handleCopy = async (e) => {
    if (e && stopPropagation) {
      e.stopPropagation();
    }
    const copyOutput = copy(data);
    if (copyOutput) {
      setCopySuccess(true);
    }
  };

  useEffect(() => {
    let timeoutId;
    if (copySuccess) {
      timeoutId = setTimeout(() => {
        setCopySuccess(false);
      }, 1000);
    }
    return () => clearTimeout(timeoutId);
  }, [copySuccess]);

  useEffect(() => {
    const triggerCopy = async () => {
      // @ts-ignore
      await handleCopy();
      setDoCopy(false);
    };
    if (doCopy) {
      triggerCopy();
    }
  }, [doCopy, setDoCopy]);

  return (
    <>
      {!copySuccess && (
        <CopyToClipboardIcon
          className={classNames("h-6 w-6 cursor-pointer", className)}
          onClick={handleCopy}
        />
      )}
      {copySuccess && (
        <CopyToClipboardSuccessIcon className="h-6 w-6 text-ok" />
      )}
    </>
  );
};

export default CopyToClipboard;

export { CopyToClipboardProvider };
