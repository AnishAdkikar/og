from transformers import BertTokenizer, BertModel
import torch
import pandas as pd
import sys
# Load the dataset from a CSV file
def load_dataset(file_path):
    # Replace this with the actual loading code, e.g.:
    # df = pd.read_csv(file_path)
    # For demonstration, we will create a DataFrame manually
    df = pd.DataFrame({
        'Date': ['12.10', '13.10'],
        'Time': ['09.00', '19.00'],
        'Pressure': ['120-80', '120-80'],
        'State by voice': ['tired', 'tired'],
        'Quality of sleep': ['good', 'bad'],
        'State by video': ['working', 'tired'],
        'Final state': ['good, busy', 'working tired']
    })
    return df

# Function to convert rows of a DataFrame into concatenated text
def rows_to_text(dataframe):
    return [" ".join(str(x) for x in row) for row in dataframe.values]

# Function to convert lines of text into embeddings
def text_to_embedding(texts, tokenizer, model):
    inputs = tokenizer(texts, padding=True, truncation=True, return_tensors="pt", max_length=512)
    outputs = model(**inputs)
    # Get the embeddings for the [CLS] token
    embeddings = outputs.last_hidden_state[:, 0, :].detach().numpy()
    return embeddings

# Main process function to convert dataset rows to embeddings and save to file
def process_dataset_to_embeddings(file_path, output_file):
    # Load pre-trained model and tokenizer
    tokenizer = BertTokenizer.from_pretrained('bert-base-uncased')
    model = BertModel.from_pretrained('bert-base-uncased')

    # Load the dataset
    df = load_dataset(file_path)

    # Convert dataframe rows to concatenated text
    texts = rows_to_text(df)

    # Convert texts to embeddings
    embeddings = text_to_embedding(texts, tokenizer, model)

    # Save the embeddings to a text file
    with open(output_file, 'w') as f:
        for emb in embeddings:
            # Convert each embedding to a string and write to file
            f.write(' '.join(map(str, emb)) + '\n')
    
    with open("../scripts/text_data.txt", 'w') as f:
        for i in texts:
            f.write(''.join(map(str, i)) + '\n')


# Assuming the data is in a CSV file named 'dataset.csv'
# if __name__ == "__main__":
#     input_csv_file = 'dataset.csv'  # Update this with the path to your CSV file
#     output_embeddings_file = './scripts/embeddings.txt'  # The output file where embeddings will be saved

#     process_dataset_to_embeddings(input_csv_file, output_embeddings_file)
if __name__ == "__main__":
    # Check if the correct number of command-line arguments is provided
    if len(sys.argv) != 3:
        print("Usage: python script.py <input_csv_file> or <output_txt>")
        sys.exit(1)

    # Get the input CSV file path from the command-line arguments
    input_csv_file = sys.argv[1]

    # The output file where embeddings will be saved
    output_embeddings_file = sys.argv[2]

    process_dataset_to_embeddings(input_csv_file, output_embeddings_file)