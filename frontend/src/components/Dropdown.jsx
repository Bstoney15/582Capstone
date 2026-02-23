import { useState, useRef, useEffect } from "react";
import { ChevronDownIcon } from "@heroicons/react/20/solid";
import "./Dropdown.css";

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

    useEffect(() => {
        function handleClickOutside(event) {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
                setIsOpen(false);
            }
        }
        document.addEventListener("mousedown", handleClickOutside);
        return () => {
            document.removeEventListener("mousedown", handleClickOutside);
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
