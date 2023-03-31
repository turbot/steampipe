import isEmpty from "lodash/isEmpty";
import useDeepCompareEffect from "use-deep-compare-effect";
import useTemplateRender from "../../../../hooks/useTemplateRender";
import {
  Category,
  CategoryProperties,
  KeyValuePairs,
  RowRenderResult,
} from "../../common/types";
import { classNames } from "../../../../utils/styles";
import { ErrorIcon } from "../../../../constants/icons";
import { getComponent } from "../../index";
import { useEffect, useState } from "react";

type RowPropertiesTitleProps = {
  category: Category | undefined;
  title: string;
};

type RowPropertiesProps = {
  propertySettings: CategoryProperties | null;
  properties: KeyValuePairs | null;
};

type RowPropertyItemProps = {
  name: string;
  rowTemplateData?: RowRenderResult | null;
  value: any;
  wrap: boolean;
};

const RowPropertiesTitle = ({ category, title }: RowPropertiesTitleProps) => (
  <div className="flex flex-col space-y-1">
    {category && (
      <span
        className="block text-foreground-lighter text-xs"
        style={{ color: category.color }}
      >
        {category.title || category.name}
      </span>
    )}
    <strong className="block">{title}</strong>
  </div>
);

const RowPropertyItemValue = ({
  name,
  rowTemplateData,
  value,
  wrap,
}: RowPropertyItemProps) => {
  const ExternalLink = getComponent("external_link");
  const [href, setHref] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  useEffect(() => {
    if (!rowTemplateData) {
      setHref(null);
      setError(null);
      return;
    }
    const renderedTemplateForProperty = rowTemplateData[name];
    if (!renderedTemplateForProperty) {
      setHref(null);
      setError(null);
      return;
    }
    if (renderedTemplateForProperty.result) {
      setHref(renderedTemplateForProperty.result);
      setError(null);
    } else if (renderedTemplateForProperty.error) {
      setHref(null);
      setError(renderedTemplateForProperty.error);
    }
  }, [name, rowTemplateData]);

  const wrapClassName = wrap ? "whitespace-normal" : "truncate";
  const linkClassName = classNames("link-highlight", wrapClassName);

  let cellContent;
  if (value === null || value === undefined) {
    cellContent = href ? (
      <ExternalLink to={href} className={linkClassName} title={`${name}=null`}>
        null
      </ExternalLink>
    ) : (
      <span
        className={classNames("text-foreground-lightest", wrapClassName)}
        title={`${name}=null`}
      >
        <>null</>
      </span>
    );
  } else {
    let renderValue: string;
    switch (typeof value) {
      case "object":
        renderValue = JSON.stringify(value, null, 2);
        break;
      default:
        renderValue = value.toString();
        break;
    }
    cellContent = href ? (
      <ExternalLink
        to={href}
        className={linkClassName}
        title={`${name}=${renderValue ? renderValue : "Empty"}`}
      >
        {renderValue ? renderValue : "Empty"}
      </ExternalLink>
    ) : (
      <span
        className={classNames(
          "block break-words",
          renderValue ? "" : "text-foreground-lightest",
          wrapClassName
        )}
        title={`${name}=${renderValue ? renderValue : "Empty"}`}
      >
        {renderValue ? renderValue : "Empty"}
      </span>
    );
  }

  return error ? (
    <span className="flex items-center space-x-2" title={error}>
      {cellContent} <ErrorIcon className="inline h-4 w-4 text-alert" />
    </span>
  ) : (
    cellContent
  );
};

const RowPropertyItem = ({
  name,
  value,
  rowTemplateData,
  wrap,
}: RowPropertyItemProps) => {
  return (
    <div className="w-full">
      <span
        className={classNames(
          "block text-sm text-foreground-lighter truncate",
          wrap ? "whitespace-normal" : "truncate"
        )}
        title={name}
      >
        {name}
      </span>
      <RowPropertyItemValue
        name={name}
        rowTemplateData={rowTemplateData}
        wrap={wrap}
        value={value}
      />
    </div>
  );
};

const RowProperties = ({
  properties = {},
  propertySettings,
}: RowPropertiesProps) => {
  const [rowTemplateData, setRowTemplateData] =
    useState<RowRenderResult | null>(null);
  const { ready: templateRenderReady, renderTemplates } = useTemplateRender();

  useDeepCompareEffect(() => {
    if (!templateRenderReady || isEmpty(propertySettings) || !properties) {
      setRowTemplateData(null);
      return;
    }

    const doRender = async () => {
      const templates = {};
      for (const [name, property] of Object.entries(propertySettings || {})) {
        if (property.display !== "none" && property.href) {
          templates[name] = property.href;
        }
      }
      if (isEmpty(templates) || !properties) {
        setRowTemplateData(null);
        return;
      }
      const renderedResults = await renderTemplates(templates, [properties]);
      setRowTemplateData(renderedResults[0]);
    };

    doRender();
  }, [properties, propertySettings, renderTemplates, templateRenderReady]);

  return (
    <div className="space-y-2">
      {Object.entries(properties || {}).map(([key, value]) => {
        const propertyDefinition = propertySettings?.[key];
        if (propertyDefinition && propertyDefinition.display === "none") {
          return null;
        }
        return (
          <RowPropertyItem
            key={key}
            name={key}
            value={value}
            rowTemplateData={rowTemplateData}
            wrap={
              propertyDefinition ? propertyDefinition.wrap === "all" : false
            }
          />
        );
      })}
    </div>
  );
};

export default RowProperties;

export { RowPropertiesTitle };
