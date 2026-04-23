<script lang="ts">
  import type { SchemaEvidence as Evidence } from '$lib/api/openapi/evidence';
  import { findResource, type ManufacturerInfo, type Placeholders } from '$lib/types/evidence';
  import Field from './Field.svelte';

  interface Props {
    evidences: Evidence[];
    manufacturer: ManufacturerInfo;
    placeholders: Placeholders;
  }
  let {
    evidences,
    manufacturer = $bindable(),
    placeholders = $bindable()
  }: Props = $props();

  const product = $derived(findResource(evidences, 'product'));
  const hasHardwareComponents = $derived((product?.data.hardwareIds?.length ?? 0) > 0);
  const contactPerson = $derived(findResource(evidences, 'contactPerson'));
  const application = $derived(findResource(evidences, 'application'));
  const memory = $derived(findResource(evidences, 'memory'));
  const virtualMachine = $derived(findResource(evidences, 'virtualMachine'));
  const sbomDocument = $derived(findResource(evidences, 'sbomDocument'));
  const euDeclarationOfConformity = $derived(findResource(evidences, 'euDeclarationOfConformity'));
  const cvdPolicy = $derived(findResource(evidences, 'coordinatedVulnerabilityDisclosurePolicy'));
  const distributionOfUpdates = $derived(findResource(evidences, 'distributionOfUpdatesDocument'));
  const productionAndMonitoring = $derived(findResource(evidences, 'productionAndMonitoringProcessDocument'));
  const riskAssessment = $derived(findResource(evidences, 'cyberSecurityRiskAssessmentDocument'));
  const monitoringProcedure = $derived(findResource(evidences, 'monitoringProcedure'));
</script>

<div
  class="text-[13.5px] leading-[1.8] text-gray-600 tracking-[0.01em]
    [&_h2]:text-base [&_h2]:font-semibold [&_h2]:text-gray-900 [&_h2]:mt-10 [&_h2]:mb-2 [&_h2]:tracking-[-0.015em]
    [&_h3]:text-[13.5px] [&_h3]:font-semibold [&_h3]:text-gray-800 [&_h3]:mt-7 [&_h3]:mb-1
    [&_h4]:text-[13.5px] [&_h4]:font-medium [&_h4]:text-gray-700 [&_h4]:mt-5 [&_h4]:mb-1
    [&_p]:my-1
    [&_ul]:my-1 [&_ul]:pl-5 [&_ul]:list-disc
    [&_strong]:font-semibold [&_strong]:text-gray-600"
>
  <h2>Information and Instructions to the User</h2>

  <h3>1. Manufacturer</h3>
  <p><strong>Manufacturer name:</strong> <Field bind:value={manufacturer.name} placeholder="[MANUFACTURER NAME]" /></p>
  <p><strong>Registered trade name or trademark:</strong> <Field bind:value={manufacturer.tradeName} placeholder="[TRADE NAME / TRADEMARK]" /></p>
  <p><strong>Postal address:</strong> <Field bind:value={manufacturer.postalAddress} placeholder="[POSTAL ADDRESS]" /></p>
  <p>
    <strong>General contact email:</strong>
    {#if manufacturer.generalEmail}
      <Field bind:value={manufacturer.generalEmail} placeholder="[GENERAL CONTACT EMAIL]" />
    {:else}
      <Field
        readonly
        evidence={contactPerson?.evidence ?? null}
        resourceType="contactPerson"
        field="emailAddress"
        value={contactPerson?.data.emailAddress}
        placeholder="[GENERAL CONTACT EMAIL]"
      />
    {/if}
  </p>
  <p>
    <strong>Security contact email:</strong>
    {#if manufacturer.securityEmail}
      <Field bind:value={manufacturer.securityEmail} placeholder="[SECURITY EMAIL]" />
    {:else}
      <Field
        readonly
        evidence={contactPerson?.evidence ?? null}
        resourceType="contactPerson"
        field="emailAddress"
        value={contactPerson?.data.emailAddress}
        placeholder="[SECURITY EMAIL]"
      />
    {/if}
  </p>
  <p><strong>Website:</strong> <Field bind:value={manufacturer.website} placeholder="[WEBSITE]" /></p>

  <h3>2. Vulnerability reporting</h3>
  <p>Information about vulnerabilities concerning this product may be reported to the manufacturer through the following single point of contact:</p>
  <p>
    <strong>Security contact:</strong>
    {#if manufacturer.securityEmail}
      <Field bind:value={manufacturer.securityEmail} placeholder="[SECURITY EMAIL]" />
    {:else}
      <Field
        readonly
        evidence={contactPerson?.evidence ?? null}
        resourceType="contactPerson"
        field="emailAddress"
        value={contactPerson?.data.emailAddress}
        placeholder="[SECURITY EMAIL]"
      />
    {/if}
  </p>
  <p><strong>Alternative reporting channel:</strong> <Field bind:value={manufacturer.securityPortalUrl} placeholder="[SECURITY PORTAL URL]" /></p>
  <p>
    <strong>CVD policy:</strong>
    <Field
      readonly
      evidence={cvdPolicy?.evidence ?? null}
      resourceType="coordinatedVulnerabilityDisclosurePolicy"
      field="dataLocation"
      value={cvdPolicy?.data.dataLocation?.localDataLocation?.path}
      placeholder="[CVD POLICY URL]"
    />
  </p>
  <p>When reporting a vulnerability, the reporter should, where possible, provide:</p>
  <ul>
    <li>product name and version;</li>
    <li>affected component or interface, if known;</li>
    <li>description of the suspected vulnerability;</li>
    <li>conditions under which it can be reproduced;</li>
    <li>potential impact, if known;</li>
    <li>contact details for follow-up.</li>
  </ul>
  <p>The manufacturer will handle vulnerability reports in accordance with its coordinated vulnerability disclosure policy.</p>

  <h3>3. Product identification</h3>
  <p>
    <strong>Product name:</strong>
    <Field
      readonly
      evidence={product?.evidence ?? null}
      resourceType="product"
      field="name"
      value={product?.data.name}
      placeholder="[PRODUCT NAME]"
    />
  </p>
  <p>
    <strong>Product type:</strong>
    <Field
      readonly
      evidence={product?.evidence ?? null}
      resourceType="product"
      field="type"
      value={product?.data.type}
      placeholder="[PRODUCT TYPE]"
    />
  </p>
  <p>
    <strong>Version / firmware / software version:</strong>
    <Field
      readonly
      evidence={product?.evidence ?? null}
      resourceType="product"
      field="programmingVersion"
      value={product?.data.programmingVersion}
      placeholder="[VERSION]"
    />
  </p>
  <p>
    <strong>Unique product identifier, if applicable:</strong>
    <Field
      readonly
      evidence={product?.evidence ?? null}
      resourceType="product"
      field="id"
      value={product?.data.id}
      placeholder="[SERIAL NUMBER / BATCH / DEVICE ID]"
    />
  </p>

  <h3>4. Intended purpose and essential information</h3>
  <p><strong>Intended purpose:</strong></p>
  <p>
    This product is intended for
    <Field
      readonly
      evidence={product?.evidence ?? null}
      resourceType="product"
      field="purpose"
      value={product?.data.purpose}
      placeholder="[INTENDED PURPOSE]"
    />.
  </p>
  <p><strong>Essential functionalities:</strong></p>
  <p>
    The product provides the following essential functionalities:
    <Field
      readonly
      evidence={product?.evidence ?? null}
      resourceType="product"
      field="description"
      value={product?.data.description}
      placeholder="[ESSENTIAL FUNCTIONALITIES]"
    />.
  </p>
  <p><strong>Security properties:</strong></p>
  <p>The product includes the following security-relevant properties, where applicable: authentication, access control, update capability, logging, and protection of data in storage and transmission.</p>

  <h3>5. Foreseeable circumstances of use or misuse creating cybersecurity risks</h3>
  <p>Cybersecurity risks may arise in particular where:</p>
  <ul>
    <li>the product is used outside its intended purpose;</li>
    <li>the product is operated with outdated software or firmware;</li>
    <li>default or weak credentials are used;</li>
    <li>security settings are changed without authorization or sufficient knowledge;</li>
    <li>the product is connected to untrusted networks or systems;</li>
    <li>unsupported third-party software, hardware, or services are used.</li>
  </ul>

  <h3>6. EU declaration of conformity</h3>
  <p>
    Where applicable, the EU declaration of conformity is available at:
    <Field
      readonly
      evidence={euDeclarationOfConformity?.evidence ?? null}
      resourceType="euDeclarationOfConformity"
      field="dataLocation"
      value={euDeclarationOfConformity?.data.dataLocation?.localDataLocation?.path}
      placeholder="[URL]"
    />
  </p>

  <h3>7. Technical security support</h3>
  <p><strong>Type of technical security support:</strong></p>
  <p>The manufacturer provides security-relevant updates, vulnerability handling, and security-related user support for this product.</p>
  <p>
    <strong>End date of support period: </strong>
    <Field
      readonly
      evidence={product?.evidence ?? null}
      resourceType="product"
      field="supportEnds"
      value={product?.data.supportEnds}
      placeholder="[END DATE]"
    />
  </p>
  <p>After this date, security updates or vulnerability handling for the product may no longer be provided.</p>

  <h3>8. Detailed instructions and information</h3>

  <h4>(a) Measures for secure use</h4>
  <p>For secure initial commissioning and secure use throughout the lifetime of the product, the user should:</p>
  <ul>
    <li>install the product in accordance with the manufacturer's instructions;</li>
    <li>ensure the product is updated to the latest available version before or during initial use;</li>
    <li>replace default credentials or set strong credentials where applicable;</li>
    <li>restrict access to authorized users only;</li>
    <li>use the product only in the intended operating environment;</li>
    <li>install security updates without undue delay;</li>
    <li>keep connected systems and interfaces appropriately secured.</li>
  </ul>

  <h4>(b) Effect of changes on the security of data</h4>
  <p>Changes to the product, including configuration changes, software changes, connection of external systems, or changes to user access settings, may affect the security of data. Such changes may reduce the confidentiality, integrity, or availability of data if they are not properly implemented and controlled.</p>

  <h4>(c) Installation of security-relevant updates</h4>
  <p>Security-relevant updates can be installed by <Field bind:value={placeholders.automaticUpdateMethod} placeholder="[AUTOMATIC UPDATE FUNCTION / ADMIN INTERFACE / MANUAL INSTALLATION METHOD]" />.</p>
  <p>Where applicable, the user should:</p>
  <ul>
    <li>check whether automatic security updates are enabled;</li>
    <li>verify that the product is connected to the required update source;</li>
    <li>
      follow the update instructions provided in the user interface or at
      <Field
        bind:value={placeholders.updateInstructionsUrl}
        placeholder="[UPDATE INSTRUCTIONS URL]"
      />;
    </li>
    <li>restart the product if required after installation.</li>
  </ul>
  <h4>(d) Secure decommissioning</h4>
  <p>Before disposal, resale, transfer, return, or permanent withdrawal from service, the user should:</p>
  <ul>
    <li>back up data, where necessary;</li>
    <li>remove the product from user accounts, management systems, or connected services;</li>
    <li>revoke or delete credentials, tokens, or access rights associated with the product;</li>
    <li>reset the product to factory settings or use the provided secure erase function, where available;</li>
    <li>verify that user data has been removed from the product.</li>
  </ul>
  <p><strong>Data removal method:</strong> <Field bind:value={placeholders.dataRemovalMethod} placeholder="[FACTORY RESET / SECURE WIPE / ADMIN FUNCTION / INSTRUCTIONS URL]" /></p>

  <h4>(e) Turning off automatic security updates</h4>
  <p>Where automatic installation of security updates is enabled by default, this setting can be turned off through: <Field bind:value={placeholders.disableUpdatesPath} placeholder="[SETTINGS PATH / ADMIN INTERFACE / INSTRUCTIONS URL]" /></p>
  <p>Disabling automatic security updates may increase cybersecurity risks. The manufacturer recommends keeping automatic security updates enabled.</p>

  <h4>(f) Information for integration into other products</h4>
  <p>Where this product is intended for integration into other products with digital elements, the integrator should use the product only in accordance with the integration, interface, configuration, and security instructions provided by the manufacturer.</p>
  <p>Integration information is available at: <Field bind:value={placeholders.integrationDocUrl} placeholder="[INTEGRATION DOCUMENT / URL]" /></p>
  <p>The integrator should in particular consider:</p>
  <ul>
    <li>supported interfaces and configurations;</li>
    <li>required security settings;</li>
    <li>update and patch management;</li>
    <li>access control and authentication measures;</li>
    <li>any limitations or assumptions described in the integration documentation.</li>
  </ul>

  <h2>TECHNICAL DOCUMENTATION</h2>

  <h3>1. General description of the product with digital elements</h3>

  <h4>1.1 Intended purpose</h4>
  <p>
    The product with digital elements
    <Field
      readonly
      evidence={product?.evidence ?? null}
      resourceType="product"
      field="name"
      value={product?.data.name}
      placeholder="[PRODUCT NAME]"
    />
    is intended for
    <Field
      readonly
      evidence={product?.evidence ?? null}
      resourceType="product"
      field="purpose"
      value={product?.data.purpose}
      placeholder="[INTENDED PURPOSE]"
    />.
  </p>
  <p>
    It is intended to be used in
    <Field
      readonly
      evidence={product?.evidence ?? null}
      resourceType="product"
      field="contextOfUse"
      value={product?.data.contextOfUse}
      placeholder="[INTENDED ENVIRONMENT / CONTEXT OF USE]"
    />.
  </p>

  <h4>1.2 Software versions affecting compliance</h4>
  <p>The following software / firmware versions affect compliance with the essential cybersecurity requirements:</p>
  <p>
    Product version:
    <Field
      readonly
      evidence={product?.evidence ?? null}
      resourceType="product"
      field="programmingVersion"
      value={product?.data.programmingVersion}
      placeholder="[VERSION]"
    />
  </p>
  {#if application}
    <p>
      Application:
      <Field
        readonly
        evidence={application.evidence}
        resourceType="application"
        field="name"
        value={application.data.name}
        placeholder="[NAME]"
      />
      {#if application.data.programmingLanguage || application.data.programmingVersion}
        (<Field
          readonly
          evidence={application.evidence}
          resourceType="application"
          field="programmingLanguage"
          value={application.data.programmingLanguage}
          placeholder="[LANGUAGE]"
        />
        <Field
          readonly
          evidence={application.evidence}
          resourceType="application"
          field="programmingVersion"
          value={application.data.programmingVersion}
          placeholder="[VERSION]"
        />)
      {/if}
    </p>
  {/if}
  <p>
    Software components relevant to compliance (library mit version):
    <Field
      readonly
      evidence={application?.evidence ?? null}
      resourceType="application"
      field=""
      value={undefined}
      placeholder="[NAME/VERSION]"
    />
  </p>
  <h4>1.3 External features, marking and internal layout</h4>
  {#if hasHardwareComponents}
    <p>
        Where the product is a hardware product, photographs or illustrations showing its external features, marking and internal layout are provided in:
      <Field
        bind:value={placeholders.hardwareLayoutReference}
        placeholder="[REFERENCE / ANNEX / APPENDIX]"
      />
    </p>
  {:else}
    <p>
      Not applicable. The product is a software-only product and does not contain hardware
      components, therefore photographs or illustrations of external features, marking and internal
      layout are not required.
    </p>
  {/if}

  <h4>1.4 User information and instructions</h4>
  <p>The user information and instructions for the product are provided in: </p>
  <p>Reference: <Field bind:value={placeholders.userInstructionsReference} placeholder="[DOCUMENT TITLE / REFERENCE / VERSION]" /></p>

  <h3>2. Description of design, development, production and vulnerability handling processes</h3>

  <h4>2.1 Design and development</h4>
  <p>The product has been designed and developed by <Field bind:value={manufacturer.name} placeholder="[MANUFACTURER NAME]" /> in accordance with documented design and development processes.</p>
  <p>The product consists of the following main elements:</p>
  <ul>
    <li>
      Hardware components:
      {#if memory}
        <Field
          readonly
          evidence={memory.evidence}
          resourceType="memory"
          field="name"
          value={memory.data.name}
          placeholder="[DESCRIPTION]"
        />
        {#if memory.data.mode}
          (<Field
            readonly
            evidence={memory.evidence}
            resourceType="memory"
            field="mode"
            value={memory.data.mode}
            placeholder="[MODE]"
          />)
        {/if}
      {:else}
        <Field readonly value={undefined} placeholder="[DESCRIPTION]" />
      {/if}
    </li>
    <li>
      Software / firmware components:
      {#if application}
        <Field
          readonly
          evidence={application.evidence}
          resourceType="application"
          field="name"
          value={application.data.name}
          placeholder="[DESCRIPTION]"
        />
        {#if application.data.programmingLanguage || application.data.programmingVersion}
          (<Field
            readonly
            evidence={application.evidence}
            resourceType="application"
            field="programmingLanguage"
            value={application.data.programmingLanguage}
            placeholder="[LANGUAGE]"
          />
          <Field
            readonly
            evidence={application.evidence}
            resourceType="application"
            field="programmingVersion"
            value={application.data.programmingVersion}
            placeholder="[VERSION]"
          />)
        {/if}
      {:else}
        <Field readonly value={undefined} placeholder="[DESCRIPTION]" />
      {/if}
    </li>
    <li>
      Interfaces / connections:
      {#if virtualMachine}
        <Field
          readonly
          evidence={virtualMachine.evidence}
          resourceType="virtualMachine"
          field="name"
          value={virtualMachine.data.name}
          placeholder="[DESCRIPTION]"
        />
        {virtualMachine.data.internetAccessibleEndpoint ? '(internet-accessible)' : '(not internet-accessible)'}
      {:else}
        <Field readonly value={undefined} placeholder="[DESCRIPTION]" />
      {/if}
    </li>
  </ul>
  <p>
    Where applicable, drawings, schemes and system architecture documentation are provided in:
    <Field
      bind:value={placeholders.architectureDocumentReference}
      placeholder="[REFERENCE / DOCUMENT TITLE]"
    />
  </p>

  <h4>2.2 Vulnerability handling</h4>
  <p>The manufacturer has established vulnerability handling processes for the product, including:</p>
  <ul>
    <li>receipt and assessment of vulnerability reports;</li>
    <li>documentation of relevant components;</li>
    <li>remediation and release of security updates;</li>
    <li>communication through the vulnerability contact point.</li>
  </ul>
  <p>Relevant documentation:</p>
  <ul>
    <li>
      <strong>
        Software bill of materials (SBOM):
        <Field
          readonly
          evidence={sbomDocument?.evidence ?? null}
          resourceType="sbomDocument"
          field="dataLocation"
          value={sbomDocument?.data.dataLocation?.localDataLocation?.path}
          placeholder="[REFERENCE / URL]"
        />
      </strong>
    </li>
    <li>
      <strong>
        Coordinated vulnerability disclosure policy:
        <Field
          readonly
          evidence={cvdPolicy?.evidence ?? null}
          resourceType="coordinatedVulnerabilityDisclosurePolicy"
          field="dataLocation"
          value={cvdPolicy?.data.dataLocation?.localDataLocation?.path}
          placeholder="[REFERENCE / URL]"
        />
      </strong>
    </li>
    <li>
      <strong>
      Vulnerability reporting contact:
      {#if manufacturer.securityEmail}
        <Field bind:value={manufacturer.securityEmail} placeholder="[SECURITY CONTACT EMAIL / PORTAL]" />
      {:else}
        <Field
          readonly
          evidence={contactPerson?.evidence ?? null}
          resourceType="contactPerson"
          field="emailAddress"
          value={contactPerson?.data.emailAddress}
          placeholder="[SECURITY CONTACT EMAIL / PORTAL]"
        />
      {/if}
      </strong>
    </li>
    <li>
      <strong>
        Secure distribution of updates:
        <Field
          readonly
          evidence={distributionOfUpdates?.evidence ?? null}
          resourceType="distributionOfUpdatesDocument"
          field="dataLocation"
          value={distributionOfUpdates?.data.dataLocation?.localDataLocation?.path}
          placeholder="[SHORT DESCRIPTION / REFERENCE]"
        />
      </strong>
    </li>
  </ul>

  <h4>2.3 Production, monitoring and validation</h4>
  <p>The manufacturer has established production and monitoring processes relevant to the cybersecurity of the product, including release control, monitoring, and validation activities.</p>
  <p>
    Relevant records or procedures are provided in:
    <Field
      readonly
      evidence={productionAndMonitoring?.evidence ?? null}
      resourceType="productionAndMonitoringProcessDocument"
      field="name"
      value={productionAndMonitoring?.data.name}
      placeholder="[REFERENCE / DOCUMENT TITLE]"
    />
  </p>
  {#if monitoringProcedure}
    <p>
      Monitoring interval: every
      <Field
        readonly
        evidence={monitoringProcedure.evidence}
        resourceType="monitoringProcedure"
        field="intervalMonths"
        value={monitoringProcedure.data.intervalMonths}
        placeholder="[INTERVAL]"
      />
      months
      {#if monitoringProcedure.data.name}
        (<Field
          readonly
          evidence={monitoringProcedure.evidence}
          resourceType="monitoringProcedure"
          field="name"
          value={monitoringProcedure.data.name}
          placeholder="[NAME]"
        />)
      {/if}
    </p>
  {/if}

  <h3>3. Cybersecurity risk assessment</h3>
  <p>A cybersecurity risk assessment has been carried out for the product.</p>
  <p>The assessment considers the intended purpose, intended environment, interfaces, dependencies, and reasonably foreseeable misuse of the product. It identifies relevant cybersecurity risks and the measures applied to address the applicable essential cybersecurity requirements. </p>
  <p>
    Risk assessment reference:
    <Field
      readonly
      evidence={riskAssessment?.evidence ?? null}
      resourceType="cyberSecurityRiskAssessmentDocument"
      field="name"
      value={riskAssessment?.data.name}
      placeholder="[DOCUMENT TITLE / REFERENCE / VERSION]"
    />
  </p>

  <h3>4. Information relevant to determination of the support period</h3>
  <p>
    Support period / end date:
    <Field
      readonly
      evidence={product?.evidence ?? null}
      resourceType="product"
      field="supportEnds"
      value={product?.data.supportEnds}
      placeholder="[END DATE]"
    />
  </p>

  <h3>5. Harmonised standards, common specifications, or European cybersecurity certification schemes applied</h3>
  <p>No harmonised standards, common specifications, or European cybersecurity certification schemes have been applied.</p>
</div>