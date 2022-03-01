import React, { useState, useEffect } from "react";
import copy from "copy-to-clipboard";
import {
  CopyToClipboardIcon,
  CopyToClipboardSuccessIcon,
} from "../../constants/icons";

const CopyToClipboard = ({ data, className = "text-muted" }) => {
  const [copySuccess, setCopySuccess] = useState(false);

  useEffect(() => {
    let timeoutId;
    if (copySuccess) {
      timeoutId = setTimeout(() => {
        setCopySuccess(false);
      }, 1000);
    }
    return () => clearTimeout(timeoutId);
  }, [copySuccess]);

  const handleCopy = async (event) => {
    const copyOutput = copy(data);
    if (copyOutput) {
      setCopySuccess(true);
    }
  };

  return (
    <>
      {!copySuccess && (
        <CopyToClipboardIcon
          className="h-5 w-5 cursor-pointer"
          onClick={handleCopy}
        />
      )}
      {copySuccess && (
        <CopyToClipboardSuccessIcon className="h-5 w-5 text-ok" />
      )}
    </>
  );
};

export default CopyToClipboard;
