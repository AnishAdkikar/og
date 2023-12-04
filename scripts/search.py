from embeddings import text_to_embedding
from transformers import BertTokenizer, BertModel
import json
import sys
def search_string_to_embedding(input_string):
    tokenizer = BertTokenizer.from_pretrained('bert-base-uncased')
    model = BertModel.from_pretrained('bert-base-uncased')
    embedding = text_to_embedding(input_string,tokenizer,model)
    return embedding[0]

if __name__ == "__main__":
    # Check if the correct number of command-line arguments is provided
    if len(sys.argv) != 2:
        print("Usage: python search.py <input_txt>")
        sys.exit(1)


    # The output file where embeddings will be saved
    input_string = sys.argv[1]
    result_embedding = search_string_to_embedding(input_string)

    print(json.dumps(result_embedding.tolist()))