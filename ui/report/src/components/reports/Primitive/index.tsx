import Error from "../Error";
import Placeholder from "../Placeholder";
import React from "react";

type PrimitiveWrapperProps = {
  children?: null | JSX.Element;
  error?: Error;
  ready?: boolean;
};

const Primitive = ({
  children,
  error,
  ready = true,
}: PrimitiveWrapperProps) => {
  const ErrorComponent = Error;
  const PlaceholderComponent = Placeholder.component;
  return (
    <>
      {/*primitive takes full col span of parent grid*/}
      <div className="col-span-12 m-1">
        <PlaceholderComponent animate={!!children} ready={ready || !!error}>
          <ErrorComponent error={error} />
          <>{!error ? children : null}</>
        </PlaceholderComponent>
      </div>
    </>
  );
};

export default Primitive;
