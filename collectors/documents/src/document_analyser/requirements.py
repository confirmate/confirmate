from __future__ import annotations

from dataclasses import dataclass
from typing import Dict, List, Literal, Optional

@dataclass(frozen=True)
class RequirementPrompt:
    id: str
    name: str
    prompt: str # what should the LLM look for in the document
    resource_type: Literal["genericDocument", "data"] = "genericDocument"
    response_field_name: Optional[str] = None


# Mapping table of requirement IDs to prompts
REQUIREMENTS: Dict[str, RequirementPrompt] = {
    "X.1.1.9.1": RequirementPrompt(
        id="E83",
        name="Products with digital elements shall protect the availability of essential and basic functions, also after an incident, including through resilience and mitigation measures against denial-of-service attacks",
        prompt="Business continuity policy created: describing strategies and procedures to maintain essential functions during cyberattacks and recover quickly, ensuring resilience, protecting assets.",
        response_field_name="businessContinuityPolicy"
    ),
    "X.1.1.9.2": RequirementPrompt(
        id="E81",
        name="Products with digital elements shall protect the availability of essential and basic functions, also after an incident, including through resilience and mitigation measures against denial-of-service attacks",
        prompt=(
            "Describes strategies or procedures to protect the availability of essential and basic functions of the product. And also highlights mitigation measures against denial-of-service attacks."
        ),
        response_field_name="availabilityProtection"
    ),
    "X.1.1.8": RequirementPrompt(
        id="E79",
        name="Products with digital elements shall process only data, personal or other, that are adequate, relevant and limited to what is necessary in relation to the intended purpose of the product with digital elements (data minimisation)",
        prompt=(
            "Data minimization policy created: outlining principles and practices to ensure only necessary data is collected, processed, and retained, minimizing privacy risks."
        ),
        response_field_name="dataMinimizationPolicy"
    ),
    "X.1.1.7": RequirementPrompt(
        id="E63",
        name="Products with digital elements shall protect the integrity of stored, transmitted or otherwise processed data, personal or other, commands, programs and configuration against any manipulation or modification not authorised by the user, and report on corruptions",
        prompt=(
            "Integrity protection mechanisms described: technical measure to ensure stored, transmitted and processed data are not altered or tampered with. A secure default configuration is described."
        ),
        response_field_name="integrityProtection"
    ),
    "X.1.1.6": RequirementPrompt(
        id="E63",
        name="Products with digital elements shall protect the confidentiality of stored, transmitted or otherwise processed data, personal or other, such as by encrypting relevant data at rest or in transit by state of the art mechanisms, and by using other technical means;",
        prompt=(
            "Confidentiality protection mechanisms described: technical measures such as encryption to safeguard stored, transmitted, and processed data from unauthorized access."
        ),
        response_field_name="confidentialityProtection"
    ),
    "X.1.1.5": RequirementPrompt(
        id="E67",
        name="Products with digital elements shall implement mechanisms to ensure that only authorised users have access to the product with digital elements and to data, personal or other, processed by the product with digital elements, including by providing means for user authentication and for managing user identities and access rights;",
        prompt=(
            "Access control mechanisms described: technical and organizational measures to ensure only authorized users can access the product and its data, including user authentication and identity management. (e.g. Unauthorised access detection and reporting mechanisms)"
        ),
        response_field_name="accessControlMechanisms"
    ),
    "X.1.1.4": RequirementPrompt(
        id="E61",
        name="Products with digital elements shall ensure that vulnerabilities can be addressed through security updates, including, where applicable, through automatic security updates that are installed within an appropriate timeframe enabled as a default setting, with a clear and easy-to-use opt-out mechanism, through the notification of available updates to users, and the option to temporarily postpone them;",
        prompt=(
            "Security update mechanisms described: Enabled by default and description of opt-out mechanism with instructions to temporarily postpone updates. Processes to notify users of available updates."
        ),
        response_field_name="securityUpdateMechanisms"
    ),
    "X.1.1.3": RequirementPrompt(
        id="E55",
        name="Products with digital elements shall be made available on the market with a secure by default configuration, unless otherwise agreed between manufacturer and business user in relation to a tailor-made product with digital elements, including the possibility to reset the product to its original state;",
        prompt=(
            "Secure by default configuration described, including the possibility to reset the product to its original state."
        ),
        response_field_name="secureByDefaultConfiguration"
    ),
    "X.1.1.2.1": RequirementPrompt(
        id="E51",
        name="Products with digital elements shall be made available on the market without known exploitable vulnerabilities;",
        prompt=(
            "Security testing policies described: outlining processes for identifying and addressing vulnerabilities before market release, including penetration testing, code reviews, and vulnerability assessments."
        ),
        response_field_name="securityTestingPolicies"
    ),
    "X.1.1.2.2": RequirementPrompt(
        id="E52",
        name="Products with digital elements shall be made available on the market without known exploitable vulnerabilities;",
        prompt=(
            "Documented penetration testing results: showing that the product has been tested for vulnerabilities and none were found."
        ),
        response_field_name="penetrationTestingResults"
    ),
    "X.1.1.1": RequirementPrompt(
        id="E45",
        name="Products with digital elements shall be designed, developed and produced in such a way that they ensure an appropriate level of cybersecurity based on the risks",
        prompt=(
            ""
        ),
        response_field_name="cybersecurityRiskManagement"
    ),
    "X.1.1.10": RequirementPrompt(
        id="E87",
        name="Products with digital elements shall minimise the negative impact by the products themselves or connected devices on the availability of services provided by other devices or networks;",
        prompt=(
            "Impact minimization measures described: technical and organizational strategies to reduce the negative impact of the product or connected devices on the availability of services provided by other devices or networks."
        ),
        response_field_name="impactMinimizationMeasures"
    ),
    "X.1.1.11": RequirementPrompt(
        id="E3",
        name="Products with digital elements shall be designed, developed and produced to limit attack surfaces, including external interfaces;",
        prompt=(
            "Security baseline established: defining minimum security requirements and configurations to limit attack surfaces, including external interfaces."
        ),
        response_field_name="securityBaseline"
    ),
    "X.1.1.12.1": RequirementPrompt(
        id="E93",
        name="Products with digital elements shall be designed, developed and produced to reduce the impact of an incident using appropriate exploitation mitigation mechanisms and techniques;",
        prompt=(
            "Regular data backups performed: ensuring that data can be restored in the event of a security incident, minimizing impact and downtime."
        ),
        response_field_name="dataBackups"
    ),
    "X.1.1.12.2": RequirementPrompt(
        id="E91",
        name="Products with digital elements shall be designed, developed and produced to reduce the impact of an incident using appropriate exploitation mitigation mechanisms and techniques;",
        prompt=(
            "Data Leakage Prevention (DLP) mechanisms described: technical controls implemented to prevent unauthorized data exfiltration and minimize the impact of security incidents."
        ),
        response_field_name="dataLeakagePrevention"
    ),
    "X.1.1.12.3": RequirementPrompt(
        id="E92",
        name="Products with digital elements shall be designed, developed and produced to reduce the impact of an incident using appropriate exploitation mitigation mechanisms and techniques;",
        prompt=(
            "Segregation of duties implemented: organizational measures to separate critical functions and responsibilities, reducing the risk and impact of security incidents."
        ),
        response_field_name="segregationOfDuties"
    ),
    "X.1.1.13": RequirementPrompt(
        id="E99",
        name="Products with digital elements shall provide security related information by recording and monitoring relevant internal activity, including the access to or modification of data, services or functions, with an opt-out mechanism for the user;",
        prompt=(
            "Contains log data and monitoring information: showing that the product records and monitors relevant internal activity, including access to or modification of data, services, or functions."
        ),
        resource_type="data",
        response_field_name="securityLoggingAndMonitoring"
    ),
    "X.1.1.14": RequirementPrompt(
        id="E101",
        name="Products with digital elements shall provide the possibility for users to securely and easily remove on a permanent basis all data and settings and, where such data can be transferred to other products or systems, ensure that this is done in a secure manner.",
        prompt=(
            "Secure data removal mechanisms described: providing users with the ability to permanently delete all data and settings from the product."
        ),
        response_field_name="secureDataRemoval"
    ),
    "X.1.1.14.2": RequirementPrompt(
        id="E102",
        name="Products with digital elements shall provide the possibility for users to securely and easily remove on a permanent basis all data and settings and, where such data can be transferred to other products or systems, ensure that this is done in a secure manner.",
        prompt=(
            "Secure data transfer mechanisms described: ensuring that data is transferred to other products or systems in a secure manner."
        ),
        response_field_name="secureDataTransfer"
    ),
}


def get_requirement(requirement_id: str) -> Optional[RequirementPrompt]:
    """Return a requirement prompt by ID."""
    return REQUIREMENTS.get(requirement_id)


def list_requirements() -> List[RequirementPrompt]:
    """Return all requirement prompts."""
    return list(REQUIREMENTS.values())
