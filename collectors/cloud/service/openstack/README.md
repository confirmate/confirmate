# Openstack Collector
OpenStack collector is a feature of Confirmate that retrieves information about OpenStack environments through API calls. It identifies storage, virtual machines and networks. With sufficient permissions it is also possible to collect domains and projects/tenants. Note that in OpenStack environments, projects and tenants are considered equivalent.

# Limitations
## Application Credentials
In OpenStack, application credentials are specifically created for a designated project within a specific domain. This makes the collector of domains and projects unnecessary. However, a limitation exists in that the domain ID/name and project ID/name cannot be collected without the appropriate permissions. 
*NOTE:* At the time of writing, the necessary permissions for this collector are still unknown. 

- It is not possible to add OS_TENANT_ID or OS_TENANT_NAME as environment variables. Attempting to do so will result in the following error message:
`Error authenticating with application credential: Application credentials cannot request a scope.`
- The domain ID and domain name must be provided by the environment variables: `OS_PROJECT_DOMAIN_ID` and `OS_USER_DOMAIN_NAME`