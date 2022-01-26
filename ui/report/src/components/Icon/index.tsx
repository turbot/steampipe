import React from "react";
import PropTypes from "prop-types";
import {
  FontAwesomeIcon,
  FontAwesomeIconProps,
} from "@fortawesome/react-fontawesome";
import { IconProp } from "@fortawesome/fontawesome-svg-core";
import { IconName } from "@fortawesome/fontawesome-common-types";

const propTypes = {
  icon: PropTypes.oneOfType([
    PropTypes.object,
    PropTypes.arrayOf(PropTypes.string),
    PropTypes.string,
  ]),
  rotation: PropTypes.oneOf([90, 180, 270]),
  onClick: PropTypes.func,
};

const Icon = ({
  icon,
  onClick,
  className = "",
  fixedWidth = false,
  spin = false,
  title = "",
  rotation,
  ...rest
}: FontAwesomeIconProps) => {
  const isStringIcon = icon && icon.constructor === String;

  if (
    isStringIcon &&
    (icon.startsWith("fab-") ||
      icon.startsWith("fal-") ||
      icon.startsWith("far-") ||
      icon.startsWith("fas-"))
  ) {
    const iconClass = icon.slice(0, 3);
    const iconName = icon.slice(4, icon.length);
    return (
      <FontAwesomeIcon
        className={className}
        fixedWidth={fixedWidth}
        icon={[iconClass, iconName] as IconProp}
        onClick={onClick}
        spin={spin}
        rotation={rotation}
        title={title}
      />
    );
  }

  if (!icon || (isStringIcon && icon.startsWith("fa-"))) {
    // Temp until we have all icons in place
    return (
      <FontAwesomeIcon
        className={className}
        fixedWidth={fixedWidth}
        icon={["fal", "diamond"]}
        onClick={onClick}
        spin={spin}
        rotation={rotation}
        title={title}
      />
    );
  }

  if (isStringIcon) {
    // This has no prefix, so grab the light version of it
    return (
      <FontAwesomeIcon
        className={className}
        fixedWidth={fixedWidth}
        icon={["fal", icon as IconName]}
        onClick={onClick}
        spin={spin}
        rotation={rotation}
        title={title}
      />
    );
  }

  return (
    <FontAwesomeIcon
      className={className}
      fixedWidth={fixedWidth}
      icon={icon}
      onClick={onClick}
      spin={spin}
      rotation={rotation}
      title={title}
      {...rest}
    />
  );
};

Icon.propTypes = propTypes;

export default Icon;
