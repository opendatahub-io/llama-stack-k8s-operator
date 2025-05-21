# Running LlamaStack Operator with ODH

This guide provides instructions for deploying and using LlamaStack with ODH/OpenShift AI Kserve component.

## Prerequisites

- ODH/OpenShift AI [installed](https://github.com/opendatahub-io/opendatahub-operator?tab=readme-ov-file#installation).
- LlamaStack Operator [installed](../../README.md#installation)
- Cluster configured to use GPUs (g4dn.xlarge GPU nodes recommended)

## Deploying vLLM via KServe

### 1. Setup KServe in Standard Mode

#### Configure DSCI
In the DSCI resource, set the serviceMesh managementState to Removed:

```yaml
spec:
  applicationsNamespace: redhat-ods-applications
  monitoring: opendatahub
    managementState: Managed
    namespace: opendatahub
  serviceMesh:
    controlPlane:
      metricsCollection: Istio
      name: data-science-smcp
      namespace: istio-system
    managementState: Removed
```

#### Configure DSC
In the DSC resource, configure the kserve component:

```yaml
spec:
  components:
    kserve:
      defaultDeploymentMode: RawDeployment
      RawDeploymentServiceConfig: Headed
      managementState: Managed
      serving:
        managementState: Removed
        name: knative-serving
```

Verify the setup by checking that `kserve-controller-manager` and `odh-model-controller` pods are running:

```bash
oc get pods -n opendatahub | grep -E 'kserve-controller-manager|odh-model-controller'
```

### 2. Deploy LLaMA 3.2 Model via KServe UI

1. Create a Connection:
   - Go to ODH dashboard -> Create Project -> Connections tab -> Create connection
   - **Project Name:** llamastack
   - **Connection Type:** URI - v1
   - **Connection name:** llama-3.2-3b-instruct
   - **URI:** `oci://quay.io/redhat-ai-services/modelcar-catalog:llama-3.2-3b-instruct` (5.9 GiB)

2. Deploy the Model:
   - Go to Models tab -> Single-model serving platform -> Deploy model
   - **Model deployment name:** llama-3.2-3b-instruct
   - **Serving runtime:** vLLM NVIDIA GPU ServingRuntime for KServe
   - **Model server size:** Custom (1 CPU core, 10 GiB memory)
   - **Accelerator:** NVIDIA GPU (1)
   - **Additional serving runtime arguments:**
     ```
     --dtype=half
     --max-model-len=20000
     --gpu-memory-utilization=0.95
     --enable-chunked-prefill
     --enable-auto-tool-choice
     --tool-call-parser=llama3_json
     --chat-template=/app/data/template/tool_chat_template_llama3.2_json.jinja
     ```

3. Verify the model deployment:
```bash
oc get inferenceservice -n llamastack
oc get pods -n llamastack | grep llama
```

### 3. Create LlamaStackDistribution CR

```yaml
apiVersion: llamastack.io/v1alpha1
kind: LlamaStackDistribution
metadata:
  name: llamastack-custom-distribution
  namespace: llamastack
spec:
  replicas: 1
  server:
    containerSpec:
      env:
        - name: INFERENCE_MODEL
          value: llama-32-3b-instruct
        - name: VLLM_URL
          value: 'http://llama-32-3b-instruct-predictor.llamastack.svc.cluster.local:80/v1'
        - name: MILVUS_DB_PATH
          value: /.llama/distributions/remote-vllm/milvus.db
      name: llama-stack
      port: 8321
    distribution:
      image: 'quay.io/redhat-et/llama:vllm-0.2.6'
      podOverrides:
        volumeMounts:
          - mountPath: /root/.llama
            name: llama-storage
        volumes:
          - emptyDir: {}
            name: llama-storage
```

### 4. Verify LlamaStack Deployment

Check the status of the LlamaStackDistribution:
```bash
oc get llamastackdistribution -n llamastack
```

Check the running pods:
```bash
oc get pods -n llamastack | grep llamastack-custom-distribution
```

Check the logs of the LlamaStack pod:
```bash
oc logs -n llamastack -l app=llama-stack
```

Expected log output should include:
```
INFO: Started server process
INFO: Waiting for application startup.
INFO: Application startup complete.
INFO: Uvicorn running on http://['::', '0.0.0.0']:8321
```

### 5. Query the Model from Jupyter Notebook

1. Go to ODH dashboard -> Workbenches -> Workbench -> Minimal Python

2. Install required packages:
```bash
pip install llama_stack
```

3. Initialize the client:
```python
from llama_stack_client import Agent, AgentEventLogger, RAGDocument, LlamaStackClient
client = LlamaStackClient(base_url="http://llamastack-custom-distribution-service.llamastack.svc.cluster.local:8321")
```

4. Verify model registration:
```python
client.models.list()
```

Expected output should show both the LLamA model and the embedding model:
```
[Model(identifier='llama-32-3b-instruct', metadata={}, api_model_type='llm',
       provider_id='vllm-inference', provider_resource_id='llama-32-3b-instruct',
       type='model', model_type='llm'),
 Model(identifier='all-MiniLM-L6-v2', metadata={'embedding_dimension': 384.0},
       api_model_type='embedding', provider_id='sentence-transformers',
       provider_resource_id='all-MiniLM-L6-v2', type='model', model_type='embedding')]
```
