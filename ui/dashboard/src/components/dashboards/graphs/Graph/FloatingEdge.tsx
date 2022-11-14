import RowProperties from "./RowProperties";
import Tooltip from "./Tooltip";
import { classNames } from "../../../../utils/styles";
import { circleGetBezierPath, getEdgeParams } from "./utils";
import { EdgeLabelRenderer, useStore } from "reactflow";
import { useCallback } from "react";

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

  const edgeLabel = (
    <div>
      <p
        className="block p-1 bg-dashboard-panel text-black-scale-4 italic max-w-[70px] text-sm text-center text-wrap line-clamp-2"
        title={label}
      >
        {label}
      </p>
    </div>
  );

  return (
    <>
      <path
        id={id}
        d={d}
        markerEnd={markerEnd}
        style={{
          ...(style || {}),
          stroke: color || themeColors.blackScale4,
          strokeWidth: 1,
        }}
      />
      <EdgeLabelRenderer>
        <div
          className={classNames(
            "absolute pointer-events-auto",
            row_data?.properties ? "cursor-pointer" : null
          )}
          style={{
            transform: `translate(-50%, -50%) translate(${mx}px,${my}px)`,
          }}
        >
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
              {edgeLabel}
            </Tooltip>
          )}
          {(!row_data || !row_data.properties) && edgeLabel}
        </div>
      </EdgeLabelRenderer>
    </>
  );
};

export default FloatingEdge;
