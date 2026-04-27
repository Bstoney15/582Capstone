// Author: Benjamin Stonestreet
// Created: 2026-02-20
// Description: This component is a dropdown menu for the application.

import { useState, useRef, useEffect } from "react";
import { ChevronDownIcon } from "@heroicons/react/20/solid";
import "./Dropdown.css";

/**
 * A customizable dropdown menu component.
 * @param {Object} props
 * @param {Array} props.options List of options to display.
 * @param {Object} props.selectedItem The currently selected item.
 * @param {Function} props.onSelect Callback invoked when an option is selected.
 * @param {string} props.placeholder Text to display when no option is selected.
 * @param {string} props.labelKey Property key to use for the display label.
 * @param {string} props.valueKey Property key to use for the unique identifier.
 * @returns {JSX.Element} The rendered Dropdown.
 */
export default function Dropdown({
    options = [],
    selectedItem = null,
    onSelect = () => { },
    placeholder = "Select an option...",
    labelKey = "name",
    valueKey = "id"
}) {
    const [isOpen, setIsOpen] = useState(false);
    const dropdownRef = useRef(null);

    // Effect to close the dropdown when clicking outside of it
    useEffect(() => {
        /**
         * Handles clicks outside the dropdown to close it.
         * @param {MouseEvent} event The mouse event.
         */
        function handleClickOutside(event) {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
                setIsOpen(false);
            }
        }
        document.addEventListener("mousedown", handleClickOutside);
        return () => {
            document.removeEventListener("mousedown", handleClickOutside); // Cleanup listener
        };
    }, []);

    return (
        <div className="custom-dropdown-container" ref={dropdownRef}>
            <button
                className="custom-dropdown-trigger"
                onClick={() => setIsOpen(!isOpen)}
                aria-haspopup="listbox"
                aria-expanded={isOpen}
            >
                <span className="custom-dropdown-label">
                    {selectedItem ? selectedItem[labelKey] : placeholder}
                </span>
                <ChevronDownIcon className={`custom-dropdown-icon ${isOpen ? 'open' : ''}`} aria-hidden="true" />
            </button>

            {isOpen && (
                <div className="custom-dropdown-menu" role="listbox">
                    {options.map((option, index) => {
                        const isSelected = selectedItem && selectedItem[valueKey] === option[valueKey];
                        return (
                            <button
                                key={option[valueKey] || index}
                                role="option"
                                aria-selected={isSelected}
                                className={`custom-dropdown-item ${isSelected ? 'active' : ''}`}
                                onClick={() => {
                                    onSelect(option);
                                    setIsOpen(false);
                                }}
                            >
                                {option[labelKey]}
                            </button>
                        );
                    })}
                </div>
            )}
        </div>
    );
}
