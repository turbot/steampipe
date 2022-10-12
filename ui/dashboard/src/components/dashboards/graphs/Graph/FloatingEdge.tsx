import RowProperties from "./RowProperties";
import Tooltip from "./Tooltip";
import { classNames } from "../../../../utils/styles";
import { circleGetBezierPath, getEdgeParams } from "./utils";
import { useCallback } from "react";
import { useStore } from "react-flow-renderer";

const FloatingEdge = ({
  id,
  source,
  target,
  markerEnd,
  style,
  data: { color, fields, row_data, label, themeColors },
}) => {
  // const edgeLabelRef = useRef(null);
  const sourceNode = useStore(
    useCallback((store) => store.nodeInternals.get(source), [source])
  );
  const targetNode = useStore(
    useCallback((store) => store.nodeInternals.get(target), [target])
  );

  if (!sourceNode || !targetNode) {
    return null;
  }

  const { sx, sy, mx, my, tx, ty, sourcePos, targetPos } = getEdgeParams(
    sourceNode,
    targetNode
  );

  //const d = getSimpleBezierPath({
  const d = circleGetBezierPath({
    sourceX: sx,
    sourceY: sy,
    sourcePosition: sourcePos,
    targetPosition: targetPos,
    targetX: tx,
    targetY: ty,
  });

  const edge = (
    <g className="relative">
      <path
        id={id}
        // className="stroke-[1] stroke-foreground-light"
        className="react-flow__edge-path"
        d={d}
        markerEnd={markerEnd}
        style={{
          ...(style || {}),
          stroke: color || themeColors.blackScale3,
          strokeWidth: 1,
        }}
      />
      <foreignObject
        className="z-50"
        height="32"
        width="60"
        x={mx - 30}
        y={my - 16}
        requiredExtensions="http://www.w3.org/1999/xhtml"
      >
        <div
          className={classNames(
            "h-full flex items-center align-center",
            row_data?.properties ? "cursor-context-menu" : null
          )}
        >
          <p
            className="mx-auto px-1 inline-block text-center bg-dashboard-panel text-black-scale-4 italic text-sm text-wrap leading-4 line-clamp-2"
            title={label}
          >
            {label}
          </p>
        </div>
      </foreignObject>
      {/*<EdgeText*/}
      {/*  // className="italic text-[5px] fill-foreground-light"*/}
      {/*  // className="react-flow__edge-text"*/}
      {/*  // className="text-alert"*/}
      {/*  x={mx}*/}
      {/*  y={my}*/}
      {/*  label={label}*/}
      {/*  labelStyle={labelStyle}*/}
      {/*  labelShowBg={labelShowBg}*/}
      {/*  labelBgStyle={labelBgStyle}*/}
      {/*  labelBgPadding={labelBgPadding}*/}
      {/*  labelBgBorderRadius={labelBgBorderRadius}*/}
      {/*/>*/}
    </g>
  );

  return (
    <>
      {row_data && row_data.properties && (
        <Tooltip
          overlay={
            <RowProperties
              fields={fields || null}
              properties={row_data.properties}
            />
          }
          title={label}
        >
          {edge}
        </Tooltip>
      )}
      {(!row_data || !row_data.properties) && edge}
    </>
  );
};

export default FloatingEdge;
