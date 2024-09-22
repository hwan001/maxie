import React from "react";

const DropdownItem = ({ icon, label, path }) => {
  return (
    <div className="dropdown-item">
      {/*item.submenu.img ? <img src={submenuItem.img}/> : <></>*/}
      <i className={`fa-brands fa-${icon}`}></i>
      <a href={path}>
        <div>{label}</div>
      </a>
    </div>
  );
};

export default DropdownItem;
