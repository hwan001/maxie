import React from "react";
import "../styles/Button.css";
import { Link } from "react-router-dom";

const STYLES = ["btn--primary", "btn--outline", "btn-mobile--outline"];
const SIZES = [
	"btn--medium",
	"btn--large",
	"btn-mobile--small",
	"btn-mobile--medium",
];

export const Button = ({
	children,
	type,
	onClick,
	buttonStyle,
	buttonSize,
	linkTo,
}) => {
	// buttonstyle이 따로 지정되지 않으면 가장 기본인 배열의 0번째가 설정된다.
	const checkButtonStyle = STYLES.includes(buttonStyle)
		? buttonStyle
		: STYLES[0];
	// buttonsize가 지정되지 않으면 가장 기본 사이즈인 0번재 btn-medium이 설정된다.
	const checkButtonSize = SIZES.includes(buttonSize) ? buttonSize : SIZES[0];

	return (
		<Link to={linkTo} className="link-mobile">
			<button
				className={`btn ${checkButtonStyle} ${checkButtonSize}`}
				onClick={onClick}
				type={type}
			>
				{children}
			</button>
		</Link>
	);
};
