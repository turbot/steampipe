const PanelTitle = ({ name, title }) => {
  if (!name || !title) {
    return null;
  }
  return (
    <h3 id={`${name}-title`} className="truncate" title={title}>
      {title}
    </h3>
  );
};

export default PanelTitle;
