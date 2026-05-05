#!/usr/bin/env python3
"""
BSI C3A (Cloud Computing Autonomy) Catalog Importer

Downloads the BSI C3A PDF and converts it to a Confirmate catalog JSON.
"""

import json
import re
import sys
import urllib.request
from pathlib import Path
from typing import Optional

try:
    from pypdf import PdfReader
except ImportError:
    print("Error: pypdf is required. Install with: pip install pypdf")
    sys.exit(1)

BSI_C3A_URL = "https://www.bsi.bund.de/SharedDocs/Downloads/EN/BSI/Publications/CloudComputing/C3A_Cloud_Computing_Autonomy.pdf"
BSI_C3A_URL_PARAMS = "?__blob=publicationFile&v=6"


def download_pdf(url: str, output_path: Path) -> None:
    """Download the PDF from the BSI website."""
    full_url = url + BSI_C3A_URL_PARAMS if not url.endswith("?") else url
    print(f"Downloading {full_url}...")
    urllib.request.urlretrieve(full_url, output_path)
    print(f"Downloaded to {output_path}")


def extract_control_id(text: str) -> Optional[tuple[str, str, str]]:
    """Extract control ID and name from a line like '2.1 SOV-1 Strategic Sovereignty'."""
    # Match patterns like "2.1 SOV-1 Strategic Sovereignty" or "2.1.1 SOV-1-01 Jurisdiction"
    match = re.match(r'[\d.]+\s+(SOV-\d+)\s+(.+)$', text.strip())
    if match:
        return match.group(1), match.group(2).strip()
    return None


def extract_criterion_id(text: str) -> Optional[str]:
    """Extract criterion ID like SOV-1-01-C1 or SOV-1-01-AC."""
    match = re.search(r'(SOV-\d+-\d+-[CA])\d*\s+Criterion', text)
    if match:
        return match.group(1)
    match = re.search(r'(SOV-\d+-\d+-[CA])\d*\s+Additional', text)
    if match:
        return match.group(1)
    match = re.search(r'(SOV-\d+-\d+-[CA])\s+Criterion', text)
    if match:
        return match.group(1)
    match = re.search(r'(SOV-\d+-\d+-[CA])\s+Additional', text)
    if match:
        return match.group(1)
    return None


def parse_bsi_c3a(pdf_path: Path) -> dict:
    """Parse the BSI C3A PDF and extract the catalog structure."""
    reader = PdfReader(str(pdf_path))
    
    categories = []
    current_category = None
    current_control = None
    last_control_id_prefix = None
    
    for page_num, page in enumerate(reader.pages, 1):
        text = page.extract_text()
        lines = text.split('\n')
        
        for line in lines:
            line = line.strip()
            
            # Skip empty lines or lines that are just numbers
            if not line or re.match(r'^[\d\s.]+$', line):
                continue
            
            # Detect main category (e.g., "2.1 SOV-1 Strategic Sovereignty")
            # Main categories: SOV-1, SOV-2, SOV-3, SOV-4, SOV-5, SOV-6
            cat_match = re.match(r'^(\d+\.\d+)\s+(SOV-[1-6])\s+(.+)$', line)
            if cat_match and any(keyword in line for keyword in ['Strategic', 'Legal', 'Data', 'Operational', 'Supply', 'Technology']):
                cat_id = cat_match.group(2)
                cat_name = cat_match.group(3).strip()
                # Clean up name (remove trailing dots and page numbers)
                cat_name = re.sub(r'\.+\s*\d+$', '', cat_name).strip()
                cat_name = re.sub(r'\.+$', '', cat_name).strip()
                cat_name = f"{cat_id} {cat_name}"
                
                # Only create a new category if the previous one has controls or we don't have one
                if current_category and current_category.get("controls"):
                    categories.append(current_category)
                
                # Check if we already have this category
                existing = [c for c in categories if c["name"] == cat_name]
                if existing:
                    # Use existing category
                    current_category = existing[0]
                else:
                    current_category = {
                        "name": cat_name,
                        "description": f"BSI Cloud Computing Autonomy criteria - {cat_name}",
                        "controls": []
                    }
                current_control = None
                continue
            
            # Detect sub-control (e.g., "2.1.1 SOV-1 Jurisdiction")
            # Format: number.number.number SOV-X Name (e.g., "2.1.1 SOV-1 Jurisdiction")
            sub_match = re.match(r'^(\d+\.\d+\.\d+)\s+(SOV-\d+)\s+(.+)$', line)
            if sub_match and current_category:
                # The control ID is like SOV-1, we need to add the number: SOV-1-01
                # From "2.1.1 SOV-1 Jurisdiction", we get "SOV-1" and the suffix from 2.1.1 = 01
                section_num = sub_match.group(1)  # e.g., "2.1.1"
                sov_base = sub_match.group(2)  # e.g., "SOV-1"
                control_name = sub_match.group(3).strip()
                control_name = re.sub(r'\.+$', '', control_name).strip()
                
                # Extract the last number from section (2.1.1 -> 01)
                suffix = section_num.split('.')[-1]
                control_id = f"{sov_base}-{suffix.zfill(2)}"
                
                if control_name:
                    last_control_id_prefix = control_id
                    current_control = {
                        "id": control_id,
                        "name": control_name,
                        "description": f"BSI C3A: {control_id} - {control_name}",
                        "controls": []
                    }
                    current_category["controls"].append(current_control)
                continue
            
            # Detect criterion (e.g., "SOV-1-01-C1 Criterion")
            if ('Criterion' in line or 'Additional' in line) and last_control_id_prefix:
                # Find all criteria IDs in this line
                crit_ids = re.findall(r'(SOV-\d+-\d+-[CA]\d*)', line)
                for criterion_id in crit_ids:
                    criterion_desc = line.strip()
                    criterion_desc = re.sub(r'^\d+\.\d+\.\d+\s+', '', criterion_desc)
                    criterion_desc = re.sub(r'\.+$', '', criterion_desc).strip()
                    
                    sub_control = {
                        "id": criterion_id,
                        "description": criterion_desc,
                        "metrics": []
                    }
                    if current_control:
                        current_control["controls"].append(sub_control)
    
    if current_category and current_category.get("controls"):
        categories.append(current_category)
    
    return {
        "id": "BSI_C3A",
        "name": "BSI C3A Cloud Computing Autonomy",
        "description": "Criteria enabling Cloud Computing Autonomy (C3A) by the German Federal Office for Information Security (BSI). These criteria are based on the EU Cloud Sovereignty Framework and C5:2026.",
        "short_name": "C3A",
        "all_in_scope": False,
        "categories": categories
    }


def main():
    import argparse
    
    parser = argparse.ArgumentParser(description="BSI C3A Catalog Importer")
    parser.add_argument("--download", action="store_true", help="Download the PDF first")
    parser.add_argument("--pdf-path", default="/tmp/bsi-c3a.pdf", help="Path to the PDF file")
    parser.add_argument("--output", default="example-data/catalogs/bsi-c3a-catalog.json", help="Output JSON path")
    
    args = parser.parse_args()
    
    pdf_path = Path(args.pdf_path)
    
    if args.download or not pdf_path.exists():
        download_pdf(BSI_C3A_URL, pdf_path)
    
    print("Parsing PDF...")
    catalog = parse_bsi_c3a(pdf_path)
    
    output_path = Path(args.output)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    
    print(f"Writing catalog to {output_path}...")
    with open(output_path, 'w') as f:
        json.dump([catalog], f, indent=2)
    
    print(f"Done! Created catalog with {len(catalog['categories'])} categories.")


if __name__ == "__main__":
    main()